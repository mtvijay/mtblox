package main

import (
	"net/http"
)

const (
	COMMON_NAME_URL     = "[a-zA-Z0-9][a-zA-Z0-9_.-]{2,255}"
	COMMON_NAME_PATTERN = "[a-zA-Z0-9][a-zA-Z0-9_.-]+"
	COMMON_ID_URL       = "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}"
)

func initProjectHandlers() HttpRouteLayout {
        m := HttpRouteLayout{
                Prefix: "/api/v1/projects",
                Handlers: HttpApiHandleArray{
                        "POST": {
                                "": HttpApiFunc{projectCreate, "json"},
                        },
                        "GET": {
                                "": HttpApiFunc{projectList, "json"},
                                "/id/{prgId:" + COMMON_ID_URL + "}":   HttpApiFunc{tagReadById, "jsonpb"},
                                "/name/{prgName:" + COMMON_NAME_URL + "}":  HttpApiFunc{tagReadByName, "jsonpb"},
                        },
                        "PUT": {
                                "/id/{prgId:" + COMMON_ID_URL + "}": HttpApiFunc{tagUpdate, "jsonpb"},
                        },
                        "DELETE": {
                                "/id/{prgId:" + COMMON_ID_URL + "}": HttpApiFunc{tagDelete, "jsonpb"},
                        },
                        "OPTIONS": {
                                "": HttpApiFunc{optionsHandler, ""},
                        },
                },
        }
        return m
}


func optionsHandler(ctx *Mcontext) (int, interface{}) {
	return http.StatusOK, nil
}

func projectCreate(ctx *Mcontext) (int, interface{}) {
	return 400, nil
}

func projectList(ctx *Mcontext) (int, interface{}) {
	return 400, nil
}

func tagReadByName(ctx *Mcontext) (int, interface{}) {
	return 400, nil
}

func tagReadById(ctx *Mcontext) (int, interface{}) {
	return 400, nil
}

func tagUpdate(ctx *Mcontext) (int, interface{}) {
	return 400, nil
}

func tagDelete(ctx *Mcontext) (int, interface{}) {
	return 400, nil
}
