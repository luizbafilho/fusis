---
title: Scheduling Modes
permalink: /scheduling-modes
layout: default
---

# Scheduling Modes

This page describes the algorithms for allocating TCP connections and UDP datagrams to real servers.
All these options are the ones available in the IPVS Kernel module.

Option | Name | Description
:--- | :--- | :---
**rr** | Round Robin | Distributes jobs equally amongst the available real servers.
**wrr** | Weighted Round Robin | Assigns jobs to real servers proportionally to there real servers' weight. Servers with higher  weights  receive new jobs first and get more jobs than servers with lower weights. Servers with equal weights get an equal distribution of new jobs.
**lc** | Least-Connection | Assigns more jobs to real servers with fewer active jobs.
**wlc** | Weighted Least-Connection | Assigns more jobs to servers with fewer jobs and relative to the real servers' weight (Ci/Wi). This is the default.
**lblc** | Locality-Based Least-Connection | Assigns jobs destined for the same IP address to the same server if the server is not overloaded and available; otherwise assign jobs to servers with fewer jobs, and keep it for future assignment.
**lblcr** | Locality-Based Least-Connection with Replication | Assigns jobs destined for the same IP address to the least-connection node in the server set for the IP address. If all the node in the server set are over loaded, it picks up a node with fewer jobs in the cluster and adds it in the sever set for the target. If the server set has not been modified for the specified time, the most loaded node is removed from the server set, in order to avoid high degree of replication.
**dh** | Destination Hashing | Assigns jobs to servers through looking up a statically assigned hash table by their destination IP addresses.
**sh** | Source Hashing | Assigns jobs to servers through looking up a statically assigned hash table by their source IP addresses.
**sed** | Shortest Expected Delay | Assigns an incoming job to the server with the shortest expected delay. The expected delay that the job will experience is (Ci + 1) / Ui if sent to the ith server, in which Ci is the number of jobs on the the ith server and Ui is the fixed service rate (weight) of the ith server.
**nq** | Never Queue | Assigns an incoming job to an idle server if there is, instead of waiting for a fast one; if all the servers are busy, it adopts the Shortest Expected Delay policy to assign the job.
