package main

import "fmt"

func main() {
	keyStr := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDLZDFn5KXgxFbHp6TNGr43TISslxBxUigoyqzeJj4g77LzMSh3tfsRHdHxFfqjQyzUBMJYd6cj0U87G1e2cEg98aHz/8PSmPFaUcNUWwszl/RdWD8yaJA8Qy9ZhvyLa7HQLdGcKRSM9mRwcIDEJQx31YV98yjZOygmZ/QmyYvYhwIDAQAB"
	key := getSenzieRsaPub(keyStr)
	if key != nil {
		println("hooooo")
	}

	sig := "gI1g3e8GluSQVxHv8ztEgAHYA2DIDovXvtMobW9vzOPpI+ZeTGMnank3/YZG9exI9Zr5jo1UH+UyuVOfML00a3IoMd1n+Nr6mGVlHaNRXE0bBnG0LIZrR6eV3b58JMb/Q/3/CsAl31puvv+JD43FPOq/Hfe5IrpC0ZMdL7GTAK8="
	msg := "SHARE #pubkey MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDLZDFn5KXgxFbHp6TNGr43TISslxBxUigoyqzeJj4g77LzMSh3tfsRHdHxFfqjQyzUBMJYd6cj0U87G1e2cEg98aHz/8PSmPFaUcNUWwszl/RdWD8yaJA8Qy9ZhvyLa7HQLdGcKRSM9mRwcIDEJQx31YV98yjZOygmZ/QmyYvYhwIDAQAB #uid 15261116645893SiMPd1F5pynK6Ka2o79WP9QCx9g #time 1526111664589 @senzswitch ^3SiMPd1F5pynK6Ka2o79WP9QCx9g"

	err := verify(msg, sig, key)
	if err != nil {
		fmt.Println("Error sign:", err.Error())
	}
}
