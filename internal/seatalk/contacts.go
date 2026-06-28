package seatalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const employeeCodeURL = "https://openapi.seatalk.io/contacts/v2/get_employee_code_with_email"

type employeeCodeResp struct {
	Code      int `json:"code"`
	Employees []struct {
		Code           int    `json:"code"`
		Email          string `json:"email"`
		EmployeeCode   string `json:"employee_code"`
		EmployeeStatus int    `json:"employee_status"`
	} `json:"employees"`
}

// EmployeeCodesByEmail 批量把 email 换成 employee_code(单次最多 500),返回 email -> employee_code;
// 查不到的 email 不会出现在结果里。同一 email 可能对应多个员工(不同在职状态),优先取在职(status=2)。
func (c *Client) EmployeeCodesByEmail(emails []string) (map[string]string, error) {
	if len(emails) == 0 {
		return map[string]string{}, nil
	}
	token, err := c.accessToken()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(map[string][]string{"emails": emails})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, employeeCodeURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get employee code: unexpected status %d", res.StatusCode)
	}

	var resp employeeCodeResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("get employee code: response code %d", resp.Code)
	}

	out := make(map[string]string)
	chosenStatus := make(map[string]int) // email -> 已选中员工的在职状态,用于优先在职
	for _, e := range resp.Employees {
		if e.Code != 0 || e.EmployeeCode == "" {
			continue
		}
		// 已选中在职(2)的,不再被非在职覆盖
		if prev, ok := chosenStatus[e.Email]; ok && prev == 2 && e.EmployeeStatus != 2 {
			continue
		}
		out[e.Email] = e.EmployeeCode
		chosenStatus[e.Email] = e.EmployeeStatus
	}
	return out, nil
}
