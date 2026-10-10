package main

import "fmt"

func (a *api) mailRule(r *request) (any, error) {
	table, columns, notFound := "virtual_aliases", "id,source,destination", 29
	switch r.parts[0] {
	case "bcc":
		table, columns, notFound = "recipient_bcc", "id,source,destination,region", 32
	case "transport":
		table, columns, notFound = "transport_domains", "id,source,destination,region", 33
	}
	domainID := r.parts[1]
	if r.Method == "GET" {
		if len(r.parts) == 2 {
			return a.list(r, "SELECT "+columns+" FROM "+table+" WHERE domain_id = $1", domainID)
		}
		return a.one(r, notFound, "SELECT "+columns+" FROM "+table+" WHERE id = $1 AND domain_id = $2", r.parts[2], domainID)
	}
	if r.user.level < 5 {
		return nil, apiError(27)
	}
	if r.Method == "DELETE" {
		return nil, a.db.execute(r.Context(), "DELETE FROM "+table+" WHERE id = $1 AND domain_id = $2", r.parts[2], domainID)
	}
	obj, err := r.object()
	if err != nil {
		return nil, err
	}
	keys := []string{"source", "destination"}
	if table != "virtual_aliases" {
		keys = append(keys, "region")
	}
	if err := requireFields(obj, keys...); err != nil {
		return nil, err
	}
	source, err := stringField(obj, "source", 20)
	if err != nil {
		return nil, err
	}
	if table == "transport_domains" {
		if source != "" && !matches(transportSource, source) {
			return nil, apiError(13)
		}
	} else {
		if !matches(ruleSource, source) {
			return nil, apiError(13)
		}
		destination, err := stringField(obj, "destination", 20)
		if err != nil {
			return nil, err
		}
		if !matches(emailAddress, destination) {
			return nil, apiError(28)
		}
		source += "@"
	}
	domain, err := a.domainRow(r)
	if err != nil {
		return nil, err
	}
	source += domain["name"].(string)
	if table == "virtual_aliases" {
		return nil, a.db.execute(r.Context(), "INSERT INTO virtual_aliases(domain_id, source, destination) VALUES ($1,$2,$3)", domainID, source, obj["destination"])
	}
	return nil, a.db.execute(r.Context(), "INSERT INTO "+table+"(domain_id, source, destination, region) VALUES ($1,$2,$3,$4)", domainID, source, obj["destination"], obj["region"])
}

func (a *api) defaultTransport(r *request) (any, error) {
	if len(r.parts) == 1 {
		return map[string]any{"result": map[string]any{"1": map[string]any{
			"Illustrate": "Relay all mail to one server and store in it. This is helpful when you want to add multi servers in your mx record.",
			"parameters": `You need a json context with you server id in it. Like {"id":1}`,
		}}}, nil
	}
	if r.parts[2] != "1" {
		return nil, apiError(31)
	}
	obj, err := r.object()
	if err != nil {
		return nil, err
	}
	if err := requireFields(obj, "id"); err != nil {
		return nil, err
	}
	server, err := a.db.query(r.Context(), "SELECT * FROM my_networks WHERE id = $1", fmt.Sprint(obj["id"]))
	if err != nil {
		return nil, err
	}
	domains, err := a.db.query(r.Context(), "SELECT * FROM virtual_domains WHERE id = $1", r.parts[1])
	if err != nil {
		return nil, err
	}
	if len(server) == 0 {
		return nil, apiError(30)
	}
	if len(domains) == 0 {
		return nil, fmt.Errorf("missing domain")
	}
	return nil, a.db.transaction(r.Context(),
		statement{"DELETE FROM transport_domains WHERE domain_id = $1", []any{r.parts[1]}},
		statement{"INSERT INTO transport_domains(domain_id, source, destination, region) VALUES($1,$2,$3,$4)",
			[]any{r.parts[1], domains[0]["name"], "lmtp:unix:private/dovecot-lmtp", "0default"}},
	)
}
