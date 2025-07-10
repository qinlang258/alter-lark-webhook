package main

import (
	"fmt"
	"regexp"
)

func main() {
	url := "116981788283.dkr.ecr.ap-east-1.amazonaws.com/chief/user/chief-sso.git/sso-server:dev-dev-68c9e27e-20250710_103848"

	re := regexp.MustCompile(`116981788283.dkr.ecr.ap-east-1.amazonaws.com/(.*?).git`)

	match := re.FindStringSubmatch(url)
	fmt.Println(match[1])
}
