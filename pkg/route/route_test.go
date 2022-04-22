package route

import (
	"fmt"
	"testing"
)

func TestRule(*testing.T) {
	rules := NewRules(Service, NewRule(Any, "/api", "/api", 8433))
	base64, err := Marshal(rules)
	if err != nil {
		fmt.Println("err", err.Error())
	} else {
		fmt.Println(base64)
	}
}

func TestRoute(*testing.T) {
	route := NewRoute()

	//api
	route.Add(GET, "/api/game", "/api/game", "http://127.0.0.1:8080")

	//ws
	route.Add(POST, "/api/swagger", "/api/swagger", "http://127.0.0.1:8081")

	ip, path, ok := route.Find(GET, "/api/game/qwq/qwqeq")
	if !ok {
		fmt.Println("api/qwq/qwqeq not found")
	} else {
		fmt.Println("proxy ", ip, path)
	}

	ip, path, ok = route.Find(POST, "/api/swagger/test/ws")
	if !ok {
		fmt.Println("api/qwq/qwqeq not found")
	} else {
		fmt.Println("proxy ", ip, path)
	}

	route.Delete(GET, "/api/game")
	route.Delete(POST, "/api/swagger")

	ip, path, ok = route.Find(GET, "/api/game/qwq/qwqeq")
	if !ok {
		fmt.Println("api/qwq/qwqeq not found")
	} else {
		fmt.Println("proxy ", ip, path)
	}

	ip, path, ok = route.Find(POST, "/api/swagger/test/ws")
	if !ok {
		fmt.Println("api/qwq/qwqeq not found")
	} else {
		fmt.Println("proxy ", ip, path)
	}
}
