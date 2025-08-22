// Example modification to enable prepared statements
// In server/dbutil/db.go

// Select runs a query and reads results into dest.
func Select(ctx context.Context, db sqlx.Ext, q sq.SelectBuilder, dest interface{}) error {
	useStatement := true // Changed from false to true
	q = q.PlaceholderFormat(sq.Dollar)
	qstr, qargs, err := q.ToSql()
	if err == nil {
		if a, ok := db.(sqlx.PreparerContext); ok && useStatement {
			stmt, prepareErr := sqlx.PreparexContext(ctx, a, qstr)
			if prepareErr != nil {
				err = prepareErr
			} else {
				defer stmt.Close() // Important: close prepared statement
				err = stmt.SelectContext(ctx, dest, qargs...)
			}
		} else if a, ok := db.(sqlx.QueryerContext); ok {
			err = sqlx.SelectContext(ctx, a, dest, qstr, qargs...)
		} else {
			err = sqlx.Select(db, dest, qstr, qargs...)
		}
	}
	// ... logging code
	return err
}
