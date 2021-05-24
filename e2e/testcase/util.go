package testcase

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cloud-provider-alibaba-cloud/e2e/framework"
	req "k8s.io/cloud-provider-alibaba-cloud/pkg/controller/service"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
	"time"
)

// Test results flag
type result struct {
	slbResult bool
	listenResult bool
	testResult bool
}

var test result

func ExpectSLBExistAndEqual(m *framework.Expectation) (bool, error) {
	lbMdl := &model.LoadBalancer{
		NamespacedName: util.NamespacedName(m.Case.Service),
	}
	//0. expect slb must exist

	//init LoadBalancerMgr
	err := m.E2E.ModelBuilder.LoadBalancerMgr.Find(m.Case.ReqCtx, lbMdl)
	if err != nil {
		// TODO if err, need retry
		return false, framework.NewErrorRetry(err)
	}

	//-------------------
	if lbMdl.LoadBalancerAttribute.LoadBalancerId == ""{
		return false, framework.NewErrorRetry(err)
	}
	if lbMdl.LoadBalancerAttribute.LoadBalancerId != ""{
		framework.Logf("start test LoadBalancerAttribute %+v",lbMdl.LoadBalancerAttribute)
		slbResult,_ := SLBEqual(m, lbMdl.LoadBalancerAttribute)
		test.slbResult = slbResult
	}


	// init listenMgr
	err = m.E2E.ModelBuilder.ListenerMgr.BuildLocalModel(m.Case.ReqCtx, lbMdl)
	if err != nil {
		// TODO if err, need retry
		framework.Logf("init ListenerMgr failed retry")
		return false, framework.NewErrorRetry(err)
	}

	for _, p := range m.Case.Service.Spec.Ports {
		for _, v := range lbMdl.Listeners {
			pro := ""
			if ProtocolPort := m.Case.ReqCtx.Anno.Get(req.ProtocolPort);ProtocolPort != ""{
				framework.Logf("ProtocolPort:%+v",ProtocolPort)
				g := strings.Split(ProtocolPort,":")
				pro = g[0]
			}


			if v.Protocol == "" || string(v.ListenerPort) == "" {
				framework.Logf("get slb Protocol and ListenerPort empty: %+v\n", lbMdl.Listeners)
				return false, framework.NewErrorRetry(err)
			} else if strings.ToLower(string(p.Protocol)) == v.Protocol || pro == v.Protocol && int(p.Port) == v.ListenerPort {
				framework.Logf("Start testing configuration consistency: %+v", v)
				listenResult,_ := ListenerEqual(m, v)
				test.listenResult = listenResult


			}else {
				klog.Info("unknown err，please check!")
				framework.Logf("p---%+v\n",p)
				framework.Logf("v---%+v\n",v)
				return false, framework.NewErrorRetry(err)
			}

		}

	}
	if test.slbResult && test.listenResult{
		test.testResult = true
	}

	return test.testResult,nil
}
func ListenerEqual(m *framework.Expectation, listen model.ListenerAttribute) (done bool, err error) {
	if CertID := m.Case.ReqCtx.Anno.Get(req.CertID); CertID != "" {
		klog.Infof("expect CertID  ok:%s", CertID)
		if listen.CertId != CertID {
			klog.Info("expected: waiting slb CertId change: ", listen.CertId)
			return false, framework.NewErrorRetry(err)
		}
	}
	//slb health check
	if HealthCheckFlag := m.Case.ReqCtx.Anno.Get(req.HealthCheckFlag); HealthCheckFlag != "" {
		klog.Infof("expect listen.HealthCheck  ok:%s", HealthCheckFlag)
		if string(listen.HealthCheck) != HealthCheckFlag {
			klog.Info("expected: waiting slb HealthCheckFlag change: ", listen.HealthCheck)
			return false, framework.NewErrorRetry(err)
		}
	}

	if HealthCheckType := m.Case.ReqCtx.Anno.Get(req.HealthCheckType); HealthCheckType != "" {
		klog.Infof("expect listen.HealthCheckType  ok:%s", HealthCheckType)
		if listen.HealthCheckType != HealthCheckType {
			klog.Info("expected: waiting slb HealthCheckFlag change: ", listen.HealthCheckType)
			return false, framework.NewErrorRetry(err)
		}
	}
	if HealthCheckURI := m.Case.ReqCtx.Anno.Get(req.HealthCheckURI); HealthCheckURI != "" {
		klog.Infof("expect listen.HealthCheckURI  ok:%s", HealthCheckURI)
		if listen.HealthCheckURI != HealthCheckURI {
			klog.Info("expected: waiting slb HealthCheckURI change: ", listen.HealthCheckURI)
			return false, framework.NewErrorRetry(err)
		}
	}
	if HealthyThreshold := m.Case.ReqCtx.Anno.Get(req.HealthyThreshold); HealthyThreshold != "" {
		klog.Infof("expect listen.HealthyThreshold  ok:%s", HealthyThreshold)
		if strconv.Itoa(listen.HealthyThreshold) != HealthyThreshold {
			klog.Info("expected: waiting slb HealthyThreshold change: ", listen.HealthyThreshold)
			return false, framework.NewErrorRetry(err)
		}
	}
	if UnhealthyThreshold := m.Case.ReqCtx.Anno.Get(req.UnhealthyThreshold); UnhealthyThreshold != "" {
		klog.Infof("expect listen.UnhealthyThreshold  ok:%s", UnhealthyThreshold)
		if strconv.Itoa(listen.UnhealthyThreshold) != UnhealthyThreshold {
			klog.Info("expected: waiting slb UnhealthyThreshold change: ", listen.UnhealthyThreshold)
			return false, framework.NewErrorRetry(err)
		}
	}
	if HealthCheckTimeout := m.Case.ReqCtx.Anno.Get(req.HealthCheckTimeout); HealthCheckTimeout != "" {
		klog.Infof("expect listen.HealthCheckTimeout  ok:%s", HealthCheckTimeout)
		if strconv.Itoa(listen.HealthCheckTimeout) != HealthCheckTimeout {
			klog.Info("expected: waiting slb HealthCheckTimeout change: ", listen.HealthCheckTimeout)
			return false, framework.NewErrorRetry(err)
		}
	}
	if HealthCheckInterval := m.Case.ReqCtx.Anno.Get(req.HealthCheckInterval); HealthCheckInterval != "" {
		klog.Infof("expect listen.HealthCheckInterval  ok:%s", HealthCheckInterval)
		if strconv.Itoa(listen.HealthCheckInterval) != HealthCheckInterval {
			klog.Info("expected: waiting slb HealthCheckInterval change: ", listen.HealthCheckInterval)
			return false, framework.NewErrorRetry(err)
		}
	}
	if Scheduler := m.Case.ReqCtx.Anno.Get(req.Scheduler); Scheduler != "" {
		klog.Infof("expect listen.Scheduler  ok:%s", Scheduler)
		if string(listen.Scheduler) != Scheduler {
			klog.Info("expected: waiting slb scheduler change: ", listen.Scheduler)
			return false, framework.NewErrorRetry(err)
		}
	}
	//slb alc
	if AclID := m.Case.ReqCtx.Anno.Get(req.AclID); AclID != "" {
		klog.Infof("expect listen.AclID  ok:%s", AclID)
		if string(listen.AclId) != AclID {
			klog.Info("expected: waiting slb AclID change: ", listen.AclId)
			return false, framework.NewErrorRetry(err)
		}
	}
	if AclStatus := m.Case.ReqCtx.Anno.Get(req.AclStatus); AclStatus != "" {
		klog.Infof("expect listen.AclStatus  ok:%s", AclStatus)
		if string(listen.AclStatus) != AclStatus {
			klog.Info("expected: waiting slb AclStatus change: ", listen.AclStatus)
			return false, framework.NewErrorRetry(err)
		}
	}
	if AclType := m.Case.ReqCtx.Anno.Get(req.AclType); AclType != "" {
		klog.Infof("expect listen.AclType  ok:%s", AclType)
		if string(listen.AclType) != AclType {
			klog.Info("expected: waiting slb AclType change: ", listen.AclType)
			return false, framework.NewErrorRetry(err)
		}
	}

	return true, err
}
func SLBEqual(m *framework.Expectation, slb model.LoadBalancerAttribute) (done bool, err error) {
	//1. spec equal

	if spec := m.Case.ReqCtx.Anno.Get(req.Spec); spec != "" {

		klog.Infof("expect spec type ok: %s", spec)
		if string(slb.LoadBalancerSpec) != spec {
			klog.Info("expected: waiting slb spec change: ", slb.LoadBalancerSpec)
			return false, framework.NewErrorRetry(err)
		}
	}
	//2.network type equal
	if AddressType := m.Case.ReqCtx.Anno.Get(req.AddressType); AddressType != "" {
		klog.Infof("expect AddressType  ok: %s", AddressType)
		if string(slb.AddressType) != AddressType {
			klog.Info("expected: waiting slb AddressType change: ", slb.AddressType)
			return false, framework.NewErrorRetry(err)
		}
	}
	//3.payment type equal
	if paymentType := m.Case.ReqCtx.Anno.Get(req.ChargeType); paymentType != "" {
		klog.Infof("expect payment type ok:", paymentType)
		if string(slb.InternetChargeType) != paymentType {
			klog.Info("expected: waiting slb payment change: ", slb.InternetChargeType)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 4. LoadBalancerName equal
	if LoadBalancerName := m.Case.ReqCtx.Anno.Get(req.LoadBalancerName); LoadBalancerName != "" {
		klog.Infof("expect LoadBalancerName  ok:", LoadBalancerName)
		if string(slb.LoadBalancerName) != LoadBalancerName {
			klog.Info("expected: waiting slb LoadBalancerName change: ", slb.LoadBalancerName)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 5. VSwitchId equal
	if VSwitchId := m.Case.ReqCtx.Anno.Get(req.VswitchId); VSwitchId != "" {
		klog.Infof("expect VSwitchId  ok:", VSwitchId)
		if string(slb.VSwitchId) != VSwitchId {
			klog.Info("expected: waiting slb VswitchId change: ", slb.VSwitchId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 6. MasterZoneId equal
	if MasterZoneId := m.Case.ReqCtx.Anno.Get(req.MasterZoneID); MasterZoneId != "" {
		klog.Infof("expect MasterZoneId  ok:", MasterZoneId)
		if string(slb.MasterZoneId) != MasterZoneId {
			klog.Info("expected: waiting slb MasterZoneId change: ", slb.MasterZoneId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 7. SlaveZoneId equal
	if SlaveZoneId := m.Case.ReqCtx.Anno.Get(req.SlaveZoneID); SlaveZoneId != "" {
		klog.Infof("expect SlaveZoneId  ok:", SlaveZoneId)
		if string(slb.SlaveZoneId) != SlaveZoneId {
			klog.Info("expected: waiting slb SlaveZoneId change: ", slb.SlaveZoneId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 8. DeleteProtection equal
	if DeleteProtection := m.Case.ReqCtx.Anno.Get(req.DeleteProtection); DeleteProtection != "" {
		klog.Infof("expect DeleteProtection  ok:", DeleteProtection)
		if string(slb.DeleteProtection) != DeleteProtection {
			klog.Info("expected: waiting slb DeleteProtectionStatus change: ", slb.SlaveZoneId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 9. ModificationProtectionStatus equal
	if ModificationProtectionStatus := m.Case.ReqCtx.Anno.Get(req.ModificationProtection); ModificationProtectionStatus != "" {
		klog.Infof("expect ModificationProtectionStatus  ok:", ModificationProtectionStatus)
		if string(slb.ModificationProtectionStatus) != ModificationProtectionStatus {
			klog.Info("expected: waiting slb ModificationProtectionStatus change: ", slb.ModificationProtectionStatus)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 10. ResourceGroupId equal
	if ResourceGroupId := m.Case.ReqCtx.Anno.Get(req.ResourceGroupId); ResourceGroupId != "" {
		klog.Infof("expect ResourceGroupId  ok:", ResourceGroupId)
		if string(slb.ResourceGroupId) != ResourceGroupId {
			klog.Info("expected: waiting slb ResourceGroupId change: ", slb.ResourceGroupId)
			return false, framework.NewErrorRetry(err)
		}
	}
	// 11. Bandwidth equal
	if Bandwidth := m.Case.ReqCtx.Anno.Get(req.Bandwidth); Bandwidth != "" {
		klog.Infof("expect Bandwidth  ok:%s", Bandwidth)
		if strconv.Itoa(slb.Bandwidth) != Bandwidth {
			klog.Info("expected: waiting slb Bandwidth change: ", slb.Bandwidth)
			return false, framework.NewErrorRetry(err)
		}
	}
	return true, err
}
func EnsureDeleteSVC(m *framework.Expectation) (done bool, err error) {
	result, err := m.E2E.Client.
		CoreV1().
		Services(m.Case.Service.Namespace).Get(context.Background(), m.Case.Service.Name, metav1.GetOptions{})
	if err != nil && strings.Contains(err.Error(), "not found") {
		framework.Logf("namespace: %s service: %s delete finished .", m.Case.Service.Namespace, m.Case.Service.Name)
		return true, nil
	}
	if err == nil {
		if result.Status.String() == "Terminating" {
			framework.Logf("namespace: %s service: %s still in %s state", m.Case.Service.Namespace, m.Case.Service.Name, result.Status.String(), time.Now())

			return false, nil
		}
		err := m.E2E.Client.
			CoreV1().
			Services(m.Case.Service.Namespace).
			Delete(context.Background(), m.Case.Service.Name, metav1.DeleteOptions{})
		framework.Logf("delete service, try again from error %v", err)
		return false, nil
	}
	framework.Logf("delete service, poll error, service :%s status unknown. %s", m.Case.Service.Name, err.Error())
	return false, nil
}
