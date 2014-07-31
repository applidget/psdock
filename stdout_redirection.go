package psdock

import (
	"net/url"
)

func (p *Process) redirectStdout(eofChannel chan bool) error {
	url, err := url.Parse(p.Conf.Stdout)
	if err != nil {
		return err
	}
	logger, err := newLogger(*url, p.Conf.LogPrefix, p.Conf.LogRotation, p.StatusChannel)
	if err != nil {
		return err
	}
	go logger.startCopy(p.Pty, eofChannel)

	return nil
}
