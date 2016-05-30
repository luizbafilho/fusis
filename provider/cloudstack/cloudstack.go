package cloudstack

// const (
// 	jobInProgress = iota * 2
// 	jobFailed
// )
//
// type Cloudstack struct {
// 	ProjectId string
// 	ApiUrl    string
// 	ApiKey    string
// 	SecretKey string
// 	Hostname  string
// }
//
// func init() {
// 	provider.RegisterProviderFactory("cloudstack", newCloudstack)
// }
//
// func newCloudstack() provider.Provider {
// 	//TODO: Validate params presence
// 	hostname, _ := os.Hostname()
// 	viper.SetDefault("provider.params.hostname", hostname)
// 	return Cloudstack{
// 		ProjectId: viper.GetString("provider.params.projectId"),
// 		ApiUrl:    viper.GetString("provider.params.apiUrl"),
// 		ApiKey:    viper.GetString("provider.params.apiKey"),
// 		SecretKey: viper.GetString("provider.params.secretKey"),
// 		Hostname:  viper.GetString("provider.params.hostname"),
// 	}
// }
//
// func (i Cloudstack) SetVip() (interface{}, error) {
// 	nicId, err := i.getNicId(i.Hostname)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	ip, err := i.getIp(nicId)
// 	if err != nil {
// 		return "", err
// 	}
// 	log.Infof("====> Requested IP:", ip)
// 	i.setIpToLocalInterface(ip)
//
// 	return strings.Split(ip, "/")[0], nil
// }
//
// func (i Cloudstack) UnsetVip(ipId interface{}) error {
// 	ipid := ipId.(string)
// 	params := map[string]string{}
// 	params["id"] = ipid
// 	response := &map[string]interface{}{}
// 	err := i.doCloudStackRequest("removeIpFromNic", params, response)
// 	spew.Dump(response)
//
// 	return err
// }
//
// func (i *Cloudstack) listNicIps(vmId string) error {
// 	params := map[string]string{}
// 	params["virtualmachineid"] = vmId
// 	response := &map[string]interface{}{}
// 	err := i.doCloudStackRequest("listNics", params, response)
// 	spew.Dump(response)
// 	return err
// }
//
// func (i *Cloudstack) doCloudStackRequest(cmd string, params map[string]string, result interface{}) error {
// 	url, err := i.buildCloudStackUrl(cmd, params)
// 	if err != nil {
// 		return err
// 	}
//
// 	tr := &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}
// 	client := &http.Client{Transport: tr}
//
// 	resp, err := client.Get(url)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("Unexpected response code for %s command %d: %s", cmd, resp.StatusCode, string(body))
// 	}
// 	if result != nil {
// 		err = json.Unmarshal(body, result)
// 		if err != nil {
// 			return fmt.Errorf("Unexpected result data for %s command: %s - Body: %s", cmd, err.Error(), string(body))
// 		}
// 	}
// 	return nil
// }
//
// func (i *Cloudstack) buildCloudStackUrl(command string, params map[string]string) (string, error) {
// 	apiKey := "vr5P_5mC_H7vN1MDRQqotbW8h6EEjjnIGrDiqhLEyHJHY8lb_wznIDkeNPgjfmv45M4PCqkRX6fzxk5bMY_etQ"
// 	secretKey := "rz7-Hek8YpblTb8wOXj-oaK6ZW2sAIF_Ph7Wy53q2GLLWNrAe1px3LAGW23OW3KanOUz1OHEatLOJb1WDK8Cvw"
//
// 	params["command"] = command
// 	params["response"] = "json"
// 	params["apiKey"] = apiKey
// 	var sorted_keys []string
// 	for k := range params {
// 		sorted_keys = append(sorted_keys, k)
// 	}
// 	sort.Strings(sorted_keys)
// 	var string_params []string
// 	for _, key := range sorted_keys {
// 		queryStringParam := fmt.Sprintf("%s=%s", key, url.QueryEscape(params[key]))
// 		string_params = append(string_params, queryStringParam)
// 	}
// 	queryString := strings.Join(string_params, "&")
// 	digest := hmac.New(sha1.New, []byte(secretKey))
// 	digest.Write([]byte(strings.ToLower(queryString)))
// 	signature := base64.StdEncoding.EncodeToString(digest.Sum(nil))
// 	cloudstackUrl := "https://rjctacp.globoi.com/client/api"
//
// 	return fmt.Sprintf("%s?%s&signature=%s", cloudstackUrl, queryString, url.QueryEscape(signature)), nil
// }
//
// func (i *Cloudstack) waitForAsyncJob(jobId string) (queryAsyncJobResultResponse, error) {
// 	var jobResponse queryAsyncJobResultResponse
// 	for {
// 		err := i.doCloudStackRequest("queryAsyncJobResult", map[string]string{"jobid": jobId}, &jobResponse)
// 		if err != nil {
// 			return jobResponse, err
// 		}
// 		if jobResponse.QueryAsyncJobResultResponse.JobStatus != jobInProgress {
// 			if jobResponse.QueryAsyncJobResultResponse.JobStatus == jobFailed {
// 				return jobResponse, fmt.Errorf("job failed to complete: %#v", jobResponse.QueryAsyncJobResultResponse.JobResult)
// 			}
// 			return jobResponse, nil
// 		}
// 		time.Sleep(500 * time.Millisecond)
// 	}
// }
//
// // returns a CIDR formated IP
// func (i *Cloudstack) getIp(nicId string) (string, error) {
// 	log.Infof("====> Getting New Ip")
//
// 	params := map[string]string{}
// 	params["nicid"] = nicId
//
// 	type result struct {
// 		AddIpToVmNicResponse struct {
// 			Id    string
// 			JobId string
// 		}
// 	}
//
// 	response := &result{}
//
// 	err := i.doCloudStackRequest("addIpToNic", params, response)
// 	if err != nil {
// 		log.Error(err)
// 		return "", err
// 	}
//
// 	asyncResp, err := i.waitForAsyncJob(response.AddIpToVmNicResponse.JobId)
// 	if err != nil {
// 		log.Error(err)
// 		return "", err
// 	}
//
// 	ip := asyncResp.QueryAsyncJobResultResponse.JobResult.(map[string]interface{})["nicsecondaryip"].(map[string]interface{})["ipaddress"].(string)
// 	netId := asyncResp.QueryAsyncJobResultResponse.JobResult.(map[string]interface{})["nicsecondaryip"].(map[string]interface{})["networkid"].(string)
//
// 	mask, err := i.getNetworkMask(netId)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return fmt.Sprintf("%s/%s", ip, mask), nil
// }
//
// // It returns the mask in bits
// func (i *Cloudstack) getNetworkMask(netId string) (string, error) {
// 	log.Infof("====> Getting Network Mask")
// 	type result struct {
// 		ListNetworksResponse struct {
// 			Network []struct {
// 				Cidr string
// 			}
// 		}
// 	}
// 	params := map[string]string{}
// 	params["id"] = netId
// 	params["projectid"] = i.ProjectId
//
// 	response := &result{}
// 	err := i.doCloudStackRequest("listNetworks", params, response)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	mask := strings.Split(response.ListNetworksResponse.Network[0].Cidr, "/")[1]
// 	return mask, nil
// }
//
// func (i *Cloudstack) getNicId(name string) (string, error) {
// 	log.Infof("====> Getting NIC Id")
// 	type result struct {
// 		ListVirtualMachinesResponse struct {
// 			VirtualMachine []struct {
// 				Id  string
// 				Nic []struct {
// 					Id string
// 				}
// 			}
// 		}
// 	}
//
// 	params := map[string]string{}
// 	params["name"] = name
// 	params["projectid"] = i.ProjectId
//
// 	response := &result{}
// 	err := i.doCloudStackRequest("listVirtualMachines", params, response)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return response.ListVirtualMachinesResponse.VirtualMachine[0].Nic[0].Id, nil
// }
//
// func (i *Cloudstack) setIpToLocalInterface(ip string) error {
// 	log.Infof("====> Setting IP to local interface!")
// 	return net.AddIp(ip, "eth0")
// }
