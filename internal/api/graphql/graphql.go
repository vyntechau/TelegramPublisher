package graphql

import (
	"context"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/vyntechau/TelegramPublisher/internal/analytics"
	"github.com/vyntechau/TelegramPublisher/internal/services/marketing"
	"github.com/vyntechau/TelegramPublisher/internal/storage"
)

var newSchema = graphql.NewSchema

// Register initializes the GraphQL schema and mounts the endpoint.
func Register(mux *http.ServeMux, repo storage.Repository, analyticsSvc *analytics.Service, marketSvc *marketing.Service) error {
	// Post Type
	postType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Post",
		Fields: graphql.Fields{
			"id":                  &graphql.Field{Type: graphql.Int},
			"slug":                &graphql.Field{Type: graphql.String},
			"file_id":             &graphql.Field{Type: graphql.String},
			"file_unique_id":      &graphql.Field{Type: graphql.String},
			"file_type":           &graphql.Field{Type: graphql.String},
			"caption":             &graphql.Field{Type: graphql.String},
			"author_id":           &graphql.Field{Type: graphql.Int},
			"auto_delete_seconds": &graphql.Field{Type: graphql.Int},
			"views_count":         &graphql.Field{Type: graphql.Int},
			"likes_count":         &graphql.Field{Type: graphql.Int},
			"dislikes_count":      &graphql.Field{Type: graphql.Int},
			"is_protected":        &graphql.Field{Type: graphql.Boolean},
			"created_at":          &graphql.Field{Type: graphql.String},
		},
	})

	// User Type
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":             &graphql.Field{Type: graphql.Int},
			"telegram_id":    &graphql.Field{Type: graphql.Int},
			"username":       &graphql.Field{Type: graphql.String},
			"first_name":     &graphql.Field{Type: graphql.String},
			"role":           &graphql.Field{Type: graphql.String},
			"status":         &graphql.Field{Type: graphql.String},
			"last_active_at": &graphql.Field{Type: graphql.String},
			"created_at":     &graphql.Field{Type: graphql.String},
		},
	})

	// Report Type
	reportType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Report",
		Fields: graphql.Fields{
			"id":              &graphql.Field{Type: graphql.Int},
			"post_id":         &graphql.Field{Type: graphql.Int},
			"reported_by":     &graphql.Field{Type: graphql.Int},
			"reason":          &graphql.Field{Type: graphql.String},
			"details":         &graphql.Field{Type: graphql.String},
			"status":          &graphql.Field{Type: graphql.String},
			"resolution_note": &graphql.Field{Type: graphql.String},
			"created_at":      &graphql.Field{Type: graphql.String},
			"post":            &graphql.Field{Type: postType},
		},
	})

	// Analytics Type
	analyticsType := graphql.NewObject(graphql.ObjectConfig{
		Name: "AnalyticsSummary",
		Fields: graphql.Fields{
			"total_users":          &graphql.Field{Type: graphql.Int},
			"active_users_daily":   &graphql.Field{Type: graphql.Int},
			"active_users_weekly":  &graphql.Field{Type: graphql.Int},
			"active_users_monthly": &graphql.Field{Type: graphql.Int},
			"blocked_users":        &graphql.Field{Type: graphql.Int},
			"banned_users":         &graphql.Field{Type: graphql.Int},
			"total_posts":          &graphql.Field{Type: graphql.Int},
			"total_views":          &graphql.Field{Type: graphql.Int},
			"total_likes":          &graphql.Field{Type: graphql.Int},
			"total_dislikes":       &graphql.Field{Type: graphql.Int},
			"pending_reports":      &graphql.Field{Type: graphql.Int},
			"resolved_reports":     &graphql.Field{Type: graphql.Int},
			"total_revenue_crypto": &graphql.Field{Type: graphql.Float},
			"active_subscribers":   &graphql.Field{Type: graphql.Int},
		},
	})

	// Query Root
	queryFields := graphql.Fields{
		"analytics": &graphql.Field{
			Type: analyticsType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return analyticsSvc.GetOverview(context.Background())
			},
		},
		"posts": &graphql.Field{
			Type: graphql.NewList(postType),
			Args: graphql.FieldConfigArgument{
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
				"offset": &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				limit, _ := p.Args["limit"].(int)
				offset, _ := p.Args["offset"].(int)
				posts, _, err := repo.ListPosts(context.Background(), 0, limit, offset)
				return posts, err
			},
		},
		"post": &graphql.Field{
			Type: postType,
			Args: graphql.FieldConfigArgument{
				"slug": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				slug := p.Args["slug"].(string)
				return repo.GetPostBySlug(context.Background(), slug)
			},
		},
		"reports": &graphql.Field{
			Type: graphql.NewList(reportType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				reports, _, err := repo.ListReports(context.Background(), "", 0, 50, 0)
				return reports, err
			},
		},
		"users": &graphql.Field{
			Type: graphql.NewList(userType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				users, _, err := repo.ListUsers(context.Background(), storage.UserFilter{Limit: 50})
				return users, err
			},
		},
	}

	// Mutation Root
	mutationFields := graphql.Fields{
		"reactPost": &graphql.Field{
			Type: postType,
			Args: graphql.FieldConfigArgument{
				"post_id":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"user_id":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"reaction": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				postID := int64(p.Args["post_id"].(int))
				userID := int64(p.Args["user_id"].(int))
				reaction := p.Args["reaction"].(string)

				if err := repo.SetReaction(context.Background(), postID, userID, reaction); err != nil {
					return nil, err
				}
				return repo.GetPostByID(context.Background(), postID)
			},
		},
		"createReport": &graphql.Field{
			Type: reportType,
			Args: graphql.FieldConfigArgument{
				"post_id":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"reported_by": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"reason":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"details":     &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				details := ""
				if d, ok := p.Args["details"].(string); ok {
					details = d
				}
				rep := &storage.Report{
					PostID:     int64(p.Args["post_id"].(int)),
					ReportedBy: int64(p.Args["reported_by"].(int)),
					Reason:     p.Args["reason"].(string),
					Details:    details,
				}
				err := repo.CreateReport(context.Background(), rep)
				return rep, err
			},
		},
	}

	schema, err := newSchema(graphql.SchemaConfig{
		Query:    graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: queryFields}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields}),
	})
	if err != nil {
		return err
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	mux.Handle("/graphql", h)
	return nil
}
