package handlers

import (
	"../managers/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (t *MethodInterface) CastBallot(w http.ResponseWriter, args map[string]interface{}) (lastInsertedId int64, err error) {
	var schema string
	var dataSource string

	_, ok := args["data-source"]
	if ok {
		ds := args["data-source"].(map[string]interface{})
		schema = ds["schema"].(string)
		delete(args, "data-source")
	}

	ballot := args["ballot"].(map[string]interface{})
	sessionId := args["session_id"].(string)
	pArgs := make(map[string]interface{})

	pArgs["created_at"] = time.Now().Format(time.RFC3339)

	if len(args) > 0 {
		if _, ok := ballot["created_by"]; !ok {
			pArgs["created_by"] = "admin"
		} else {
			pArgs["created_by"] = ballot["created_by"]
		}
		if _, ok := ballot["cast_counter"]; !ok {
			pArgs["cast_counter"] = 1
		} else {
			pArgs["cast_counter"] = ballot["cast_counter"]
		}
	}

	if _, ok := ballot["election_id"]; !ok {
		return
	}
	if _, ok := ballot["ballot_id"]; !ok {
		return
	}

	pArgs["session_id"] = sessionId
	pArgs["election_id"] = ballot["election_id"]
	pArgs["ballot_id"] = ballot["ballot_id"]

	// Update candidate counter
	dataSource = "candidate_counter"

	for k, v := range ballot {
		if k == "contest" {
			for _, cv := range v.([]interface{}) {
				pArgs["contest_id"] = cv.(map[string]interface{})["contest_id"]
				for ck, conv := range cv.(map[string]interface{}) {
					if ck == "candidates" {
						for _, cand := range conv.([]interface{}) {
							candidate := cand.(map[string]interface{})
							if !candidate["selected"].(bool) {
								continue
							}
							log.Println(candidate)
							pArgs["candidate_id"] = candidate["candidate_id"]
							pArgs["cast_value"] = candidate["cast_value"]

							lastInsertedId, err = database.Create(schema, dataSource, pArgs)

							if err != nil {
								if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
									log.Fatal(e)
								}
								// TODO: implement rollback
								return
							}
						}
					}
				}
			}
		}
	}

	// Update cast counter
	dataSource = "cast_counter"
	cArgs := make(map[string]interface{})
	cArgs["session_id"] = pArgs["session_id"]
	cArgs["election_id"] = pArgs["election_id"]
	cArgs["ballot_id"] = pArgs["ballot_id"]
	cArgs["created_at"] = pArgs["created_at"]
	cArgs["created_by"] = pArgs["created_by"]

	lastInsertedId, err = database.Create(schema, dataSource, cArgs)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}

	// Unset pin code
	dataSource = "pincode"
	pcArgs := make(map[string]interface{})
	pcArgs["is_active"] = false
	pcArgs["expiration_time"] = pArgs["created_at"]
	b := []byte(fmt.Sprintf(`{"pin": "%s"}`, args["pin"].(string)))
	var k map[string]interface{}
	if err = json.Unmarshal(b, &k); err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}
	pcArgs["keys"] = k

	rowsAffected, err := database.Update(schema, dataSource, pcArgs)

	if err != nil {
		if _, e := fmt.Fprintf(w, "{\"error\": \"%v\"}", err); e != nil {
			log.Fatal(e)
		}
		return
	}

	result := fmt.Sprintf("{\"inserted\": \"%d\"}", lastInsertedId)
	if _, err := fmt.Fprintf(w, result); err != nil {
		log.Fatal(err)
	}

	return rowsAffected, err
}

