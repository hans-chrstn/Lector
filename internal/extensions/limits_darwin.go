package extensions

import "errors"

func limitWorker() error {
	return errors.New("isolated plugin workers require Linux; run Lector in Docker")
}
