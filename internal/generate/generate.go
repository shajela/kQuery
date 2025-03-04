package generate

func ExtractGroupedResult(data map[string]interface{}) *string {
	dataMap, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Step 1: Access "Get" key
	getMap, ok := dataMap["Get"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Step 2: Access "Pod" key (which is a slice)
	podSlice, ok := getMap["Pod"].([]interface{})
	if !ok || len(podSlice) == 0 {
		return nil
	}

	// Step 3: Access the first map inside "Pod" slice
	firstPod, ok := podSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	// Step 4: Access "_additional" key
	additional, ok := firstPod["_additional"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Step 5: Access "generate" key
	generate, ok := additional["generate"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Step 6: Access "groupedResult"
	groupedResult, ok := generate["groupedResult"].(string)
	if !ok {
		return nil
	}

	return &groupedResult
}
