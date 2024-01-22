package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	interfaceName := flag.String("interface", "eth0", "Network interface to sniff on")
	filter := flag.String("filter", "", "BPF filter for capturing packets")
	flag.Parse()

	// Open device
	handle, err := pcap.OpenLive(*interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Set filter if provided
	if *filter != "" {
		err = handle.SetBPFFilter(*filter)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Filter set to:", *filter)
	}

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		analyzePacket(packet)
	}
}

func analyzePacket(packet gopacket.Packet) {
	fmt.Println("----- New Packet -----")
	if netLayer := packet.NetworkLayer(); netLayer != nil {
		src, dst := netLayer.NetworkFlow().Endpoints()
		fmt.Printf("From %s to %s\n", src, dst)
	}

	if transportLayer := packet.TransportLayer(); transportLayer != nil {
		src, dst := transportLayer.TransportFlow().Endpoints()
		fmt.Printf("Transport from %s to %s\n", src, dst)
	}

	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		fmt.Printf("Application layer/Payload: %s\n", string(appLayer.Payload()))
	}

	if packet.ErrorLayer() != nil {
		fmt.Printf("Error decoding some part of the packet: %v\n", packet.ErrorLayer().Error())
	}
}
