/*
 * Nuts event octopus
 * Copyright (C) 2019. Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package pkg

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestEvent_Json(t *testing.T) {
	err := "error"
	e := Event{
		Error: &err,
	}

	jsonBytes, _ := json.Marshal(e)
	json := string(jsonBytes)

	expectedTags := []string{
		"consentId",
		"transactionId",
		"initiatorLegalEntity",
		"error",
		"externalId",
		"payload",
		"retryCount",
		"name",
		"uuid",
	}

	for _, tag := range expectedTags {
		t.Run(fmt.Sprintf("Marshalling outputs tag %s", tag), func(t *testing.T) {
			if !strings.Contains(json, tag) {
				t.Errorf("Expected json to contain tag: %s", tag)
			}
		})
	}
}
