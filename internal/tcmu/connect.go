package torustcmu

import (
	"fmt"

	"github.com/alternative-storage/go-tcmu"
	"github.com/alternative-storage/go-tcmu/scsi"
	"github.com/alternative-storage/torus/block"
	"github.com/coreos/pkg/capnslog"
)

const (
	defaultBlockSize = 4 * 1024
	devPath          = "/dev/torus"
)

var clog = capnslog.NewPackageLogger("github.com/alternative-storage/torus", "tcmu")

func ConnectAndServe(f *block.BlockFile, name string, closer chan bool) error {
	wwn := tcmu.NaaWWN{
		// TODO(barakmich): CoreOS OUI here
		OUI:      "000000",
		VendorID: tcmu.GenerateSerial(name),
	}
	// TODO(nak3): Scale up NBD by handling multiple requests.
	// Requires thread-safety across the block.BlockFile/torus.File
	//n := runtime.GOMAXPROCS(0) - 1
	n := 1

	h := &tcmu.SCSIHandler{
		HBA:        30,
		LUN:        0,
		WWN:        wwn,
		VolumeName: name,
		// 1GiB, 1K
		DataSizes: tcmu.DataSizes{
			VolumeSize: int64(f.Size()),
			BlockSize:  defaultBlockSize,
		},
		DevReady: tcmu.MultiThreadedDevReady(
			&torusHandler{
				file: f,
				name: name,
				inq: &tcmu.InquiryInfo{
					VendorID:   "AlternativeStorage",
					ProductID:  "TorusBlk",
					ProductRev: "0001",
				},
			}, n),
	}
	d, err := tcmu.OpenTCMUDevice(devPath, h)
	defer d.Close()
	if err != nil {
		return err
	}
	fmt.Printf("Attached to %s/%s. Server loop begins ... \n", devPath, name)
	<-closer
	return nil
}

type torusHandler struct {
	file *block.BlockFile
	name string
	inq  *tcmu.InquiryInfo
}

func (h *torusHandler) HandleCommand(cmd *tcmu.SCSICmd) (tcmu.SCSIResponse, error) {
	switch cmd.Command() {
	case scsi.Inquiry:
		return tcmu.EmulateInquiry(cmd, h.inq)
	case scsi.TestUnitReady:
		return tcmu.EmulateTestUnitReady(cmd)
	case scsi.ServiceActionIn16:
		return tcmu.EmulateServiceActionIn(cmd)
	case scsi.ModeSense, scsi.ModeSense10:
		return tcmu.EmulateModeSense(cmd, true)
	case scsi.ModeSelect, scsi.ModeSelect10:
		return tcmu.EmulateModeSelect(cmd, true)
	case scsi.Read6, scsi.Read10, scsi.Read12, scsi.Read16:
		return tcmu.EmulateRead(cmd, h.file)
	case scsi.Write6, scsi.Write10, scsi.Write12, scsi.Write16:
		return h.handleWrite(cmd)
	case scsi.SynchronizeCache, scsi.SynchronizeCache16:
		return h.handleSyncCommand(cmd)
	case scsi.MaintenanceIn:
		return h.handleReportDeviceID(cmd)
	default:
		clog.Debugf("Ignore unknown SCSI command 0x%x\n", cmd.Command())
	}
	return cmd.NotHandled(), nil
}
