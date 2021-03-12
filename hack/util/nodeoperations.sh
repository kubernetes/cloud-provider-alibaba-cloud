#!/bin/bash

operations='
{
			Id: "restart.kubelet",
			Retry: 3,
			Timeout: 60,
			Content: ${jjj},
},
'