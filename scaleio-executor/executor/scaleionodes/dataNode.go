package core

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	basenode "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode"
	procedural "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/procedural"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioDataNode implementation for ScaleIO Fake Node
type ScaleioDataNode struct {
	basenode.BaseScaleioNode
}

//NewData generates a Data Node object
func NewData() *ScaleioDataNode {
	myNode := &ScaleioDataNode{}
	return myNode
}

//RunStateUnknown default action for StateUnknown
func (sdn *ScaleioDataNode) RunStateUnknown() {
	reboot, err := procedural.EnvironmentSetup(state)
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	state = procedural.WaitForCleanPrereqsReboot(sdn.UpdateScaleIOState())

	errState = UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		time.Sleep(time.Duration(procedural.DelayForRebootInSeconds) * time.Second)

		rebootErr := xplatform.GetInstance().Run.Command(rebootCmdline, rebootCheck, "")
		if rebootErr != nil {
			log.Errorln("Install Kernel Failed:", rebootErr)
		}

		time.Sleep(time.Duration(procedural.WaitForRebootInSeconds) * time.Second)
	} else {
		log.Infoln("No need to reboot while installing prerequisites")
	}
}

//RunStatePrerequisitesInstalled default action for StatePrerequisitesInstalled
func (sdn *ScaleioDataNode) RunStatePrerequisitesInstalled() {
	err := procedural.NodeSetup(state)
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		continue
	}

	errState := UpdateNodeState(types.StateInstallRexRay)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInstallRexRay")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (sdn *ScaleioDataNode) RunStateInstallRexRay() {
	if state.ScaleIO.Preconfig.PreConfigEnabled {
		log.Debugln("Pre-Config is enabled skipping wait for Cluster Initialization")
	} else {
		//we need to wait because without the gateway, the rexray service restart
		//will fail
		state = procedural.WaitForClusterInitializeFinish(sdn.UpdateScaleIOState())
	}

	reboot, err := procedural.RexraySetup(state)
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		continue
	}

	err = procedural.SetupIsolator(state)
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		continue
	}

	errState := UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	state = procedural.WaitForCleanInstallReboot(sdn.UpdateScaleIOState())

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("reboot:", reboot)

		time.Sleep(time.Duration(DelayForRebootInSeconds) * time.Second)

		errState = UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		rebootErr := xplatform.GetInstance().Run.Command(rebootCmdline, rebootCheck, "")
		if rebootErr != nil {
			log.Errorln("Install Kernel Failed:", rebootErr)
		}

		time.Sleep(time.Duration(WaitForRebootInSeconds) * time.Second)
	} else {
		log.Infoln("No need to reboot while installing REX-Ray")

		errState = UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (sdn *ScaleioDataNode) RunStateSystemReboot() {
	errState := UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (sdn *ScaleioDataNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}