package errorx

import "errors"

var MAX_REFRESH_TIME = errors.New("MAX_REFRESH_TIME")
var WRONG_PARAMETER = errors.New("WRONG_PARAMETER")
var READ_PROPS_ERR = errors.New("READ_PROPS_ERR")
var MONEY_NOT_ENOUGH = errors.New("MONEY_NOT_ENOUGH")
var SET_MONEY_PROPS_ERR = errors.New("SET_MONEY_PROPS_ERR")
var CSV_CFG_EMPTY = errors.New("CSV_CFG_EMPTY")
var JSON_ERR = errors.New("JSON_ERR")
var GET_DATA_EMPTY = errors.New("GET_DATA_EMPTY")
var NOT_ENOUGH_ITEM = errors.New("NOT_ENOUGH_ITEM")
var CONVER_FAIL = errors.New("CONVER_FAIL")
var CSV_ROW_NOT_FOUND = errors.New("CSV_ROW_NOT_FOUND")

