/*
	Copyright © 2021 Miga Labs
*/
package crawler

import "github.com/sirupsen/logrus"

var (
	IpCacheSize = 400

	ModuleName = "CRAWLER"
	log        = logrus.WithField(
		"module", ModuleName,
	)
)

// common interface that all kind of crawlers need to follow
type Crawler interface {
	Help()
	Run()
	Close()
}

func Help() string {
	return "\t--config-file\tconfig-file with all the available configurations. Find an example at ./config-files/config.json"
}
