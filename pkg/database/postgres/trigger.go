package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	schemasv1alpha4 "github.com/schemahero/schemahero/pkg/apis/schemas/v1alpha4"
)

func triggerCreateStatement(trigger *schemasv1alpha4.PostgresqlTableTrigger, tableName string) (string, error) {
	triggerEventSyntax, err := triggerEvent(trigger)
	if err != nil {
		return "", errors.Wrap(err, "failed to create trigger event syntax")
	}

	o := "trigger"
	if trigger.ConstraintTrigger != nil && *trigger.ConstraintTrigger {
		o = "constraint trigger"
	}

	stmt := fmt.Sprintf(`create %s %q %s on %q`, o, trigger.Name, triggerEventSyntax, tableName)

	forEachStatement := true // pg default
	if trigger.ForEachRow != nil && *trigger.ForEachRow {
		forEachStatement = false
	}

	if forEachStatement {
		stmt = fmt.Sprintf("%s for each statement", stmt)
	} else {
		stmt = fmt.Sprintf("%s for each row", stmt)
	}

	if trigger.Condition != nil {
		stmt = fmt.Sprintf("%s when (%s)", stmt, *trigger.Condition)
	}

	stmt = fmt.Sprintf("%s execute procedure %s", stmt, trigger.ExecuteProcedure)

	return stmt, nil
}

func triggerEvent(trigger *schemasv1alpha4.PostgresqlTableTrigger) (string, error) {
	if len(trigger.Events) == 0 {
		return "", errors.New("trigger missing events")
	}

	// build the event which could be like:
	//   after insert or update of col1, col2

	// all triggers must be the same temporal event (after, before, instead of)
	temporal := ""
	if strings.HasPrefix(strings.ToLower(trigger.Events[0]), "after") {
		temporal = "after"
	} else if strings.HasPrefix(strings.ToLower(trigger.Events[0]), "before") {
		temporal = "before"
	} else if strings.HasPrefix(strings.ToLower(trigger.Events[0]), "instead of") {
		temporal = "instead of"
	} else {
		return "", errors.New("unable to parse trigger")
	}

	events := []string{}
	for _, event := range trigger.Events {
		event := strings.TrimSpace(strings.ToLower(event))

		if strings.HasPrefix(event, "after") {
			events = append(events, strings.TrimPrefix(event, "after"))
			continue
		} else if strings.HasPrefix(event, "before") {
			events = append(events, strings.TrimPrefix(event, "before"))
			continue
		} else if strings.HasPrefix(event, "instead of") {
			events = append(events, strings.TrimPrefix(event, "instead of"))
			continue
		}
	}

	return fmt.Sprintf("%s%s", temporal, strings.Join(events, " or")), nil
}
