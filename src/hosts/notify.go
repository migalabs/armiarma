package hosts

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
)

/*
	File that includes the methods to set the custom modification channels for the Libp2p host
*/

func (c *BasicLibp2pHost) standardListenF(net network.Network, addr ma.Multiaddr) {
	c.Log.Debug("Listen")
}

func (c *BasicLibp2pHost) standardListenCloseF(net network.Network, addr ma.Multiaddr) {
	c.Log.Debug("Close listen")
}

func (c *BasicLibp2pHost) standardConnectF(net network.Network, conn network.Conn) {
	c.Log.Debug("Connection")
	c.Log.Debug(fmt.Sprintf("%+v\n", conn))

	c.Log.Debug("PEers in PeerStore")
	c.Log.Debug(fmt.Printf("%+v\n", h.Network().Peerstore().Peers()))
}

func (c *BasicLibp2pHost) standardDisconnectF(net network.Network, conn network.Conn) {
	c.Log.Debug("Disconnect")
}

func (c *BasicLibp2pHost) standardOpenedStreamF(net network.Network, str network.Stream) {
	c.Log.Debug("Open Stream")
}

func (c *BasicLibp2pHost) standardClosedF(net network.Network, str network.Stream) {
	c.Log.Debug("Close")
}

//
func (c *BasicLibp2pHost) SetCustomNotifications() error {
	// generate empty bundle to set custom notifiers
	bundle := &network.NotifyBundle{
		bundle.ListenF:       c.listenF,
		bundle.ListenCloseF:  c.listenCloseF,
		bundle.ConnectedF:    c.connectedF,
		bundle.DisconnectedF: c.disconnectedF,
		bundle.OpenedStreamF: c.openedStreamF,
		bundle.ClosedStreamF: c.closedStreamF,
	}
	// read host from main struct
	h, err := c.Host()
	if err != nil {
		return err
	}
	// set the custom notifiers to the host
	h.Network().Notify(bundle)
	return nil
}
