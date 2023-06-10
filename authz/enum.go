//go:generate go run github.com/dmarkham/enumer -linecomment -type=Action,ObjectType,Relation -output=enum_strings.go
package authz

type Action int

const (
	CanView              Action = iota + 1 // can_view
	CanEdit                                // can_edit
	CanEditMembers                         // can_edit_members
	CanCreateOrg                           // can_create_org
	CanDeleteOrg                           // can_delete_org
	CanCreateFeedVersion                   // can_create_feed_version
	CanDeleteFeedVersion                   // can_delete_feed_version
	CanCreateFeed                          // can_create_feed
	CanDeleteFeed                          // can_delete_feed
)

type ObjectType int

const (
	TenantType      ObjectType = iota + 1 // tenant
	GroupType                             // org
	FeedType                              // feed
	FeedVersionType                       // feed_version
	UserType                              // user
)

type Relation int

const (
	AdminRelation   Relation = iota + 1 // admin
	MemberRelation                      // member
	ManagerRelation                     // manager
	ViewerRelation                      // viewer
	EditorRelation                      // editor
	TenantRelation                      // tenant
	ParentRelation                      // parent
)
