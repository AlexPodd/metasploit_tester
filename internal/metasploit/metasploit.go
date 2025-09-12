package metasploit

import (
	"github.com/AlexPodd/metasploit_tester_console/internal/domain"
	"github.com/fpr1m3/go-msf-rpc/rpc"
)

type MetaSploitRPC struct {
	InstanceMSF *rpc.Metasploit
	Report      *domain.Report
}

func (metaSploitRPC *MetaSploitRPC) Login(host, login, password string) (err error) {
	/*
	   metaSploitRPC.Report = &domain.Report{}
	   metaSploitRPC.InstanceMSF, err = rpc.New(host, login, password)

	   	if err != nil {
	   		log.Print(err)
	   	}

	   return err
	*/
	return nil
}

func (MetaSploitRPC *MetaSploitRPC) Run(exploitsPath []string, config domain.ConfigExploit) (*domain.Report, error) {

	return nil, nil
}

func (MetaSploitRPC *MetaSploitRPC) Reload() error {
	return nil
}
