package provider

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/vishvananda/netlink"
)

const (
	jobInProgress = iota * 2
	jobFailed
)

type CloudstackIaaS struct {
	ProjectId string
	ApiKey    string
	SecretKey string
}

type queryAsyncJobResultResponse struct {
	QueryAsyncJobResultResponse struct {
		JobStatus     int         `json:"jobstatus"`
		JobResult     interface{} `json:"jobresult"`
		JobResultType string      `json:"jobresulttype"`
		JobResultCode int         `json:"jobresultcode"`
	} `json:"queryasyncjobresultresponse"`
}

func NewCloudstackIaaS(pId, apiKey, secretKey string) *CloudstackIaaS {
	return &CloudstackIaaS{
		ProjectId: pId,
		ApiKey:    apiKey,
		SecretKey: secretKey,
	}
}

func (i *CloudstackIaaS) SetVip(vmName string) (string, error) {
	nicId, err := i.GetNicId(vmName)
	if err != nil {
		return "", err
	}

	ip, err := i.getIP(nicId)
	if err != nil {
		return "", err
	}
	log.Println("====> Requested IP:", ip)
	i.setIpToLocalInterface(ip)

	return strings.Split(ip, "/")[0], nil
}

func (i *CloudstackIaaS) DeleteVip(ipId string) error {
	params := map[string]string{}
	params["id"] = ipId
	response := &map[string]interface{}{}
	err := i.doCloudStackRequest("removeIpFromNic", params, response)
	spew.Dump(response)

	return err
}

func (i *CloudstackIaaS) ListNicIps(vmId string) error {
	params := map[string]string{}
	params["virtualmachineid"] = vmId
	response := &map[string]interface{}{}
	err := i.doCloudStackRequest("listNics", params, response)
	spew.Dump(response)
	return err
}

func (i *CloudstackIaaS) doCloudStackRequest(cmd string, params map[string]string, result interface{}) error {
	url, err := i.buildCloudStackUrl(cmd, params)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected response code for %s command %d: %s", cmd, resp.StatusCode, string(body))
	}
	if result != nil {
		err = json.Unmarshal(body, result)
		if err != nil {
			return fmt.Errorf("Unexpected result data for %s command: %s - Body: %s", cmd, err.Error(), string(body))
		}
	}
	return nil
}

func (i *CloudstackIaaS) buildCloudStackUrl(command string, params map[string]string) (string, error) {
	apiKey := "vr5P_5mC_H7vN1MDRQqotbW8h6EEjjnIGrDiqhLEyHJHY8lb_wznIDkeNPgjfmv45M4PCqkRX6fzxk5bMY_etQ"
	secretKey := "rz7-Hek8YpblTb8wOXj-oaK6ZW2sAIF_Ph7Wy53q2GLLWNrAe1px3LAGW23OW3KanOUz1OHEatLOJb1WDK8Cvw"

	params["command"] = command
	params["response"] = "json"
	params["apiKey"] = apiKey
	var sorted_keys []string
	for k := range params {
		sorted_keys = append(sorted_keys, k)
	}
	sort.Strings(sorted_keys)
	var string_params []string
	for _, key := range sorted_keys {
		queryStringParam := fmt.Sprintf("%s=%s", key, url.QueryEscape(params[key]))
		string_params = append(string_params, queryStringParam)
	}
	queryString := strings.Join(string_params, "&")
	digest := hmac.New(sha1.New, []byte(secretKey))
	digest.Write([]byte(strings.ToLower(queryString)))
	signature := base64.StdEncoding.EncodeToString(digest.Sum(nil))
	cloudstackUrl := "https://rjctacp.globoi.com/client/api"

	return fmt.Sprintf("%s?%s&signature=%s", cloudstackUrl, queryString, url.QueryEscape(signature)), nil
}

func (i *CloudstackIaaS) waitForAsyncJob(jobId string) (queryAsyncJobResultResponse, error) {
	var jobResponse queryAsyncJobResultResponse
	for {
		err := i.doCloudStackRequest("queryAsyncJobResult", map[string]string{"jobid": jobId}, &jobResponse)
		if err != nil {
			return jobResponse, err
		}
		if jobResponse.QueryAsyncJobResultResponse.JobStatus != jobInProgress {
			if jobResponse.QueryAsyncJobResultResponse.JobStatus == jobFailed {
				return jobResponse, fmt.Errorf("job failed to complete: %#v", jobResponse.QueryAsyncJobResultResponse.JobResult)
			}
			return jobResponse, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// returns a CIDR formated IP
func (i *CloudstackIaaS) getIP(nicId string) (string, error) {
	log.Println("====> Getting New Ip")

	params := map[string]string{}
	params["nicid"] = nicId

	type result struct {
		AddIpToVmNicResponse struct {
			Id    string
			JobId string
		}
	}

	response := &result{}

	err := i.doCloudStackRequest("addIpToNic", params, response)
	if err != nil {
		log.Fatalln("[Err] %v", err)
		return "", err
	}

	asyncResp, err := i.waitForAsyncJob(response.AddIpToVmNicResponse.JobId)
	if err != nil {
		log.Fatalln("[Err] %v", err)
		return "", err
	}

	ip := asyncResp.QueryAsyncJobResultResponse.JobResult.(map[string]interface{})["nicsecondaryip"].(map[string]interface{})["ipaddress"].(string)
	netId := asyncResp.QueryAsyncJobResultResponse.JobResult.(map[string]interface{})["nicsecondaryip"].(map[string]interface{})["networkid"].(string)

	mask, err := i.getNetworkMask(netId)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", ip, mask), nil
}

// It returns the mask in bits
func (i *CloudstackIaaS) getNetworkMask(netId string) (string, error) {
	log.Println("====> Getting Network Mask")
	type result struct {
		ListNetworksResponse struct {
			Network []struct {
				Cidr string
			}
		}
	}
	params := map[string]string{}
	params["id"] = netId
	params["projectid"] = i.ProjectId

	response := &result{}
	err := i.doCloudStackRequest("listNetworks", params, response)
	if err != nil {
		return "", err
	}

	mask := strings.Split(response.ListNetworksResponse.Network[0].Cidr, "/")[1]
	return mask, nil
}

func (i *CloudstackIaaS) GetNicId(name string) (string, error) {
	log.Println("====> Getting NIC Id")
	type result struct {
		ListVirtualMachinesResponse struct {
			VirtualMachine []struct {
				Id  string
				Nic []struct {
					Id string
				}
			}
		}
	}

	params := map[string]string{}
	params["name"] = name
	params["projectid"] = i.ProjectId

	response := &result{}
	err := i.doCloudStackRequest("listVirtualMachines", params, response)
	if err != nil {
		return "", err
	}

	return response.ListVirtualMachinesResponse.VirtualMachine[0].Nic[0].Id, nil
}

func (i *CloudstackIaaS) setIpToLocalInterface(ip string) error {
	log.Println("====> Setting IP to local interface!")

	link, err := netlink.LinkByName("eth0")
	if err != nil {
		return err
	}

	addr, err := netlink.ParseAddr(ip)
	if err != nil {
		return err
	}

	netlink.AddrAdd(link, addr)
	return nil
}
