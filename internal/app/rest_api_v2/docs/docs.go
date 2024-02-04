// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Simple function, that returns this REST API server health status.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "REST API server health status.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    }
                }
            }
        },
        "/health/auth/any": {
            "get": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Check if ` + "`" + `any` + "`" + ` of the two users can log in. Useful for the routes which are required by both users: regular and HA.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Check ` + "`" + `any` + "`" + ` user authentication.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/health/auth/ha": {
            "get": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Check the ` + "`" + `HA` + "`" + ` user authentication.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Check the ` + "`" + `HA` + "`" + ` user authentication.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/health/auth/regular": {
            "get": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Check the ` + "`" + `regular` + "`" + ` user authentication.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Check the ` + "`" + `regular` + "`" + ` user authentication.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/jail/all": {
            "get": {
                "description": "Get the list of all Jails, including the information about them.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jails"
                ],
                "summary": "List all Jails.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/HosterJailUtils.JailApi"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/jail/destroy/{jail_name}": {
            "delete": {
                "description": "Destroy a specific Jail using it's name as a parameter.\u003cbr\u003e` + "`" + `DANGER` + "`" + ` - destructive operation!",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jails"
                ],
                "summary": "Destroy a specific Jail.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Jail Name",
                        "name": "jail_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/jail/info/{jail_name}": {
            "get": {
                "description": "Get Jail info.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jails"
                ],
                "summary": "Get Jail info.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Jail Name",
                        "name": "jail_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/HosterJailUtils.JailApi"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/jail/start/{jail_name}": {
            "post": {
                "description": "Start a specific Jail using it's name as a parameter.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jails"
                ],
                "summary": "Start a specific Jail.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Jail Name",
                        "name": "jail_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/jail/stop/{jail_name}": {
            "post": {
                "description": "Stop a specific Jail using it's name as a parameter.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jails"
                ],
                "summary": "Stop a specific Jail.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Jail Name",
                        "name": "jail_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerSuccess"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        },
        "/jail/templates": {
            "get": {
                "description": "Get the list of all Jail templates.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Jails"
                ],
                "summary": "List all Jail templates.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerStringList"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/handlers.SwaggerError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "HosterJailUtils.JailApi": {
            "type": "object",
            "properties": {
                "backup": {
                    "type": "boolean"
                },
                "config_file_append": {
                    "type": "string"
                },
                "cpu_limit_percent": {
                    "type": "integer"
                },
                "current_host": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "dns_server": {
                    "type": "string"
                },
                "encrypted": {
                    "type": "boolean"
                },
                "ip_address": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "network": {
                    "type": "string"
                },
                "parent": {
                    "type": "string"
                },
                "production": {
                    "type": "boolean"
                },
                "ram_limit": {
                    "type": "string"
                },
                "release": {
                    "type": "string"
                },
                "running": {
                    "type": "boolean"
                },
                "shutdown_script": {
                    "type": "string"
                },
                "space_free_b": {
                    "type": "integer"
                },
                "space_free_h": {
                    "type": "string"
                },
                "space_used_b": {
                    "type": "integer"
                },
                "space_used_h": {
                    "type": "string"
                },
                "startup_script": {
                    "type": "string"
                },
                "timezone": {
                    "type": "string"
                },
                "uptime": {
                    "type": "string"
                }
            }
        },
        "handlers.SwaggerError": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "handlers.SwaggerStringList": {
            "type": "object",
            "properties": {
                "message": {
                    "description": "success",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "handlers.SwaggerSuccess": {
            "type": "object",
            "properties": {
                "message": {
                    "description": "success",
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "",
	BasePath:         "/api/v2",
	Schemes:          []string{},
	Title:            "Hoster Node REST API Docs",
	Description:      "REST API documentation for the `Hoster` nodes. This HTTP endpoint is located directly on the `Hoster` node.<br>Please, take some extra care with the things you execute here, because many of them can be destructive and non-revertible (e.g. vm destroy, snapshot rollback, host reboot, etc).",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
