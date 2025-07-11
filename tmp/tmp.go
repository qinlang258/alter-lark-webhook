package main

import (
	"fmt"
	"regexp"
)

func main() {
	//message := make(map[string]string)

	imageUrl := "116981788283.dkr.ecr.ap-east-1.amazonaws.com/chief-web/chief_fe_webapp.git/chief-fe-webapp:allenv-master-dab34715-20250710_191904"

	//fieldsList := strings.Split(imageUrl, ":")
	//commitId := strings.Split(fieldsList[1], "-")[3]

	imageRe := regexp.MustCompile(`116981788283.dkr.ecr.ap-east-1.amazonaws.com/(.*?).git/([^:]+)`)

	match := imageRe.FindStringSubmatch(imageUrl)

	// message["projectPath"] = match[1]
	// message["serviceName"] = match[2]
	fmt.Println("imageUrl: ", imageUrl)
	fmt.Println(match)
}
