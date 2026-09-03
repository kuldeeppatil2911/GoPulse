package graphql

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	graphql1 "gopulse/internal/graphql"
)

type Resolver struct{}

// ServiceMetrics is the resolver for the serviceMetrics field.
func (r *queryResolver) ServiceMetrics(ctx context.Context, serviceID string, timeRange *graphql1.TimeRange) (*graphql1.ServiceMetrics, error) {
	panic("not implemented")
}

// EndpointPerformance is the resolver for the endpointPerformance field.
func (r *queryResolver) EndpointPerformance(ctx context.Context, endpointID string, timeRange *graphql1.TimeRange) ([]*graphql1.EndpointPerformance, error) {
	panic("not implemented")
}

// ErrorStatistics is the resolver for the errorStatistics field.
func (r *queryResolver) ErrorStatistics(ctx context.Context, serviceID string, timeRange *graphql1.TimeRange) ([]*graphql1.ErrorStatistics, error) {
	panic("not implemented")
}

// APIUsage is the resolver for the apiUsage field.
func (r *queryResolver) APIUsage(ctx context.Context, serviceID string, timeRange *graphql1.TimeRange) ([]*graphql1.APIUsage, error) {
	panic("not implemented")
}

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql1.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
