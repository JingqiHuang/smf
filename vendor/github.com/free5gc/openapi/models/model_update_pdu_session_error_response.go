/*
 * Nsmf_PDUSession
 *
 * SMF PDU Session Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package models

type UpdatePduSessionErrorResponse struct {
	JsonData               *HsmfUpdateError `json:"jsonData,omitempty" multipart:"contentType:application/json"`
	BinaryDataN1SmInfoToUe []byte           `json:"binaryDataN1SmInfoToUe,omitempty" multipart:"contentType:application/vnd.3gpp.5gnas,ref:JsonData.N1SmInfoToUe.ContentId"`
}
