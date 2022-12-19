package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/auth/repo/groups/db"
	"github.com/mainflux/mainflux/pkg/errors"
)

var errCreateMetadataQuery = errors.New("failed to create query for metadata")

var _ db.QueryFramer = (*postgresQueryFramer)(nil)

type postgresQueryFramer struct{}

func New() db.QueryFramer {
	return &postgresQueryFramer{}
}

func (pqf *postgresQueryFramer) Save(g auth.Group) string {
	// For root group path is initialized with id
	q := `INSERT INTO groups (name, description, id, path, owner_id, metadata, created_at, updated_at)
		  VALUES (:name, :description, :id, :id, :owner_id, :metadata, :created_at, :updated_at)
		  RETURNING id, name, owner_id, parent_id, description, metadata, path, nlevel(path) as level, created_at, updated_at`
	if g.ParentID != "" {
		// Path is constructed in insert_group_tr - init.go
		q = `INSERT INTO groups (name, description, id, owner_id, parent_id, metadata, created_at, updated_at)
			 VALUES ( :name, :description, :id, :owner_id, :parent_id, :metadata, :created_at, :updated_at)
			 RETURNING id, name, owner_id, parent_id, description, metadata, path, nlevel(path) as level, created_at, updated_at`
	}
	return q
}

func (pqf *postgresQueryFramer) Update() string {
	q := `UPDATE groups SET name = :name, description = :description, metadata = :metadata, updated_at = :updated_at WHERE id = :id
	RETURNING id, name, owner_id, parent_id, description, metadata, path, nlevel(path) as level, created_at, updated_at`
	return q
}

func (pqf *postgresQueryFramer) Delete() string {
	qd := `DELETE FROM groups WHERE id = :id`

	return qd
}

func (pqf *postgresQueryFramer) RetrieveByID() string {
	q := `SELECT id, name, owner_id, parent_id, description, metadata, path, nlevel(path) as level, created_at, updated_at FROM groups WHERE id = $1`

	return q
}

func (pqf *postgresQueryFramer) RetrieveAllChildren() (string, string) {
	q := `SELECT g.id, g.name, g.owner_id, g.parent_id, g.description, g.metadata, g.path,  nlevel(g.path) as level, g.created_at, g.updated_at
	FROM groups parent, groups g
	WHERE parent.id = :id AND g.path <@ parent.path AND nlevel(g.path) - nlevel(parent.path) < :level`

	cq := `SELECT COUNT(*) FROM groups parent, groups g WHERE parent.id = :id AND g.path <@ parent.path `

	return q, cq
}

func (pqf *postgresQueryFramer) RetrieveAllParents() (string, string) {

	q := `SELECT g.id, g.name, g.owner_id, g.parent_id, g.description, g.metadata, g.path, nlevel(g.path) as level, g.created_at, g.updated_at
FROM groups parent, groups g
WHERE parent.id = :id AND g.path @> parent.path AND nlevel(parent.path) - nlevel(g.path) <= :level`
	cq := `SELECT COUNT(*) FROM groups parent, groups g WHERE parent.id = :id AND g.path @> parent.path`

	return q, cq
}

func (pqf *postgresQueryFramer) RetrieveAll(pm auth.PageMetadata) (string, string, error) {
	_, metaQuery, err := getGroupsMetadataQuery("groups", pm.Metadata)
	if err != nil {
		return "", "", err
	}

	var mq string
	if metaQuery != "" {
		mq = fmt.Sprintf(" AND %s", metaQuery)
	}

	q := fmt.Sprintf(`SELECT id, owner_id, parent_id, name, description, metadata, path, nlevel(path) as level, created_at, updated_at FROM groups
					  WHERE nlevel(path) <= :level %s ORDER BY path`, mq)
	cq := "SELECT COUNT(*) FROM groups"
	if metaQuery != "" {
		cq = fmt.Sprintf(" %s WHERE %s", cq, metaQuery)
	}
	return q, cq, nil

}

func (pqf *postgresQueryFramer) Members(groupType string, pm auth.PageMetadata) (string, string, error) {

	_, mq, err := getGroupsMetadataQuery("groups", pm.Metadata)
	if err != nil {
		return "", "", err
	}

	q := fmt.Sprintf(`SELECT gr.member_id, gr.group_id, gr.type, gr.created_at, gr.updated_at FROM group_relations gr
					  WHERE gr.group_id = :group_id AND gr.type = :type %s`, mq)

	if groupType == "" {
		q = fmt.Sprintf(`SELECT gr.member_id, gr.group_id, gr.type, gr.created_at, gr.updated_at FROM group_relations gr
		                 WHERE gr.group_id = :group_id %s`, mq)
	}

	qc := fmt.Sprintf(`SELECT COUNT(*) FROM groups g, group_relations gr
	WHERE gr.group_id = :group_id AND gr.group_id = g.id AND gr.type = :type %s;`, mq)

	return q, qc, nil

}

func (pqf *postgresQueryFramer) Memberships(pm auth.PageMetadata) (string, string, error) {

	_, mq, err := getGroupsMetadataQuery("groups", pm.Metadata)
	if err != nil {
		return "", "", err
	}
	if mq != "" {
		mq = fmt.Sprintf("AND %s", mq)
	}
	q := fmt.Sprintf(`SELECT g.id, g.owner_id, g.parent_id, g.name, g.description, g.metadata
					  FROM group_relations gr, groups g
					  WHERE gr.group_id = g.id and gr.member_id = :member_id
		  			  %s ORDER BY id LIMIT :limit OFFSET :offset;`, mq)

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM group_relations gr, groups g
						WHERE gr.group_id = g.id and gr.member_id = :member_id %s `, mq)
	return q, cq, nil
}

func (pqf *postgresQueryFramer) Assign() string {
	q := `INSERT INTO group_relations (group_id, member_id, type, created_at, updated_at)
	VALUES(:group_id, :member_id, :type, :created_at, :updated_at)`
	return q
}

func (pqf *postgresQueryFramer) Unassign() string {
	q := `DELETE from group_relations WHERE group_id = :group_id AND member_id = :member_id`
	return q
}

func getGroupsMetadataQuery(db string, m auth.GroupMetadata) (mb []byte, mq string, err error) {
	if len(m) > 0 {
		mq = `metadata @> :metadata`
		if db != "" {
			mq = db + "." + mq
		}

		b, err := json.Marshal(m)
		if err != nil {
			return nil, "", errors.Wrap(err, errCreateMetadataQuery)
		}
		mb = b
	}
	return mb, mq, nil
}
