package main

import (
)

type errorMsg struct {
	HttpStatusMsg string
	HttpStatusCode int
}

func ZsrvMethodNotAllowed() errorMsg {
	mterr := errorMsg{}

	return mterr
}

func ZsrvUnauthorized() errorMsg {
	mterr := errorMsg{}

	return mterr
}

func ZsrvForbidden() errorMsg {
	mterr := errorMsg{}

	return mterr
}

func ZsrvObjNotFound() errorMsg {
	mterr := errorMsg{}

	return mterr
}
