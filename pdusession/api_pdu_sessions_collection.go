// Copyright 2019 free5GC.org
//
// SPDX-License-Identifier: Apache-2.0

/*
 * Nsmf_PDUSession
 *
 * SMF PDU Session Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package pdusession

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PostPduSessions - Create
func PostPduSessions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
