package engine

import (
	"fmt"
	"log"

	connPrep "github.com/ri2q/goRacer/internal/preparer/connection"
	"golang.org/x/net/http2"
)

/*
StartStream initializes the HTTP/2 connection (client preface, SETTINGS, initial WINDOW_UPDATE)
and starts a read loop that acknowledges SETTINGS, tracks stream completion, and coordinates shutdown.
*/
func StartStream(
	st *SessionState,
	connCfg *connPrep.ConnConfig,
	maxStreams uint32,
	initWindow uint32,
	headerTable uint32,
) error {

	//connCfg.TcpConn.SetNoDelay(false)
	if err := connCfg.EnableDelay(); err != nil {
		return fmt.Errorf("Error from StartStream: %v", err)
	}

	st.Framer = http2.NewFramer(connCfg.TlsConn, connCfg.TlsConn)

	// Magic Frame
	_, err := connCfg.TlsConn.Write([]byte(http2.ClientPreface))
	if err != nil {
		return fmt.Errorf("Error sending ClientPreface: %v", err)
	}

	// Settings Frame
	if err := st.Framer.WriteSettings(
		http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: maxStreams},
		http2.Setting{ID: http2.SettingInitialWindowSize, Val: initWindow},
		http2.Setting{ID: http2.SettingHeaderTableSize, Val: headerTable},
	); err != nil {
		return fmt.Errorf("Error sending settingsframe: %v", err)
	}
	// Connection wide WINDOW_UPDATE
	if err := st.Framer.WriteWindowUpdate(0, 15663105); err != nil {
		return fmt.Errorf("Error sending windowUpdate: %v", err)
	}
	// false until we received GOAWAY from the sender.
	*st.ShutDownReady = false
	if err := connCfg.DisableDelay(); err != nil {
		return fmt.Errorf("Error from StartStream: %v", err)
	}
	log.Println("Preface, Settings, WindowUpdate SID 0 DONE.")
	return nil
}
