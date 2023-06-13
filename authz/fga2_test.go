package authz

var fgaTestData = []fgaTestTuple{
	{
		Subject:  NewEntityKey(TenantType, "1"),
		Object:   NewEntityKey(GroupType, "1"),
		Relation: ParentRelation,
		Notes:    "org:test-org",
	},
	{
		Subject:  NewEntityKey(TenantType, "1"),
		Object:   NewEntityKey(GroupType, "2"),
		Relation: ParentRelation,
		Notes:    "org:restricted-org",
	},
	{
		Subject:  NewEntityKey(TenantType, "1"),
		Object:   NewEntityKey(GroupType, "3"),
		Relation: ParentRelation,
		Notes:    "org:all-member",
	},
	{
		Subject:  NewEntityKey(TenantType, "1"),
		Object:   NewEntityKey(GroupType, "4"),
		Relation: ParentRelation,
		Notes:    "org:admins-only",
	},
	{
		Subject:  NewEntityKey(TenantType, "1#member"),
		Object:   NewEntityKey(GroupType, "3"),
		Relation: ViewerRelation,
	},
	{
		Subject:  NewEntityKey(TenantType, "2"),
		Object:   NewEntityKey(GroupType, "5"),
		Relation: ParentRelation,
		Notes:    "org:no-one",
	},
	{
		Subject:  NewEntityKey(GroupType, "1"),
		Object:   NewEntityKey(FeedType, "1"),
		Relation: ParentRelation,
		Notes:    "feed:1 should be viewable to members of org:1 (ian drew) and editable by org:1 editors (drew)",
	},
	{
		Subject:  NewEntityKey(GroupType, "2"),
		Object:   NewEntityKey(FeedType, "2"),
		Relation: ParentRelation,
		Notes:    "feed:2 should be viewable to members of org:2 () and editable by org:2 editors (ian)",
	},
	{
		Subject:  NewEntityKey(GroupType, "3"),
		Object:   NewEntityKey(FeedType, "3"),
		Relation: ParentRelation,
		Notes:    "feed:3 should be viewable to all members of tenant:1 (admin nisar ian drew) and editable by org:3 editors ()",
	},
	{
		Subject:  NewEntityKey(GroupType, "4"),
		Object:   NewEntityKey(FeedType, "4"),
		Relation: ParentRelation,
		Notes:    "feed:4 should only be viewable to admins of tenant:1 (admin)",
	},
	{
		Subject:  NewEntityKey(FeedType, "2"),
		Object:   NewEntityKey(FeedVersionType, "1"),
		Relation: ParentRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "admin"),
		Object:   NewEntityKey(TenantType, "1"),
		Relation: AdminRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "ian"),
		Object:   NewEntityKey(GroupType, "1"),
		Relation: ViewerRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "ian"),
		Object:   NewEntityKey(GroupType, "2"),
		Relation: EditorRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "ian"),
		Object:   NewEntityKey(TenantType, "1"),
		Relation: MemberRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "drew"),
		Object:   NewEntityKey(GroupType, "1"),
		Relation: EditorRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "drew"),
		Object:   NewEntityKey(TenantType, "1"),
		Relation: MemberRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "nisar"),
		Object:   NewEntityKey(TenantType, "1"),
		Relation: MemberRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "nisar"),
		Object:   NewEntityKey(FeedVersionType, "1"),
		Relation: ViewerRelation,
	},
	{
		Subject:  NewEntityKey(UserType, "test2"),
		Object:   NewEntityKey(TenantType, "2"),
		Relation: MemberRelation,
	},
}

var fgaTests = []fgaTestTuple{}
