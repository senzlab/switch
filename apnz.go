package main

import (
	"fmt"
	"log"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

func main() {
	cert, err := certificate.FromP12File("apn.p12", "")
	if err != nil {
		log.Fatal("Cert Error:", err)
	}
	client := apns2.NewClient(cert).Development()

	// notification
	notification := &apns2.Notification{}
	notification.DeviceToken = "FF497B297252FB5EE007B7FD41961068B7CD8798431AE3F0CF799E003F5BB280"
	notification.Topic = "com.creative.igift"
	//notification.Payload = []byte(`{"aps":{"alert":"Hello!"}}`)
	payload := payload.NewPayload().Alert("New iGift").Badge(1).Custom("senz", "SHARE #acc 223 #amnt 23232 #cid 3223 @tes ^eran ewwe2ee2323=")
	notification.Payload = payload

	res, err := client.Push(notification)
	if err != nil {
		log.Fatal("Error:", err)
	}

	fmt.Printf("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
}
