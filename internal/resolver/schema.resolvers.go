package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"foodworks.ml/m/internal/generated/ent"
	"foodworks.ml/m/internal/generated/ent/car"
	"foodworks.ml/m/internal/generated/ent/user"
	generated "foodworks.ml/m/internal/generated/graphql"
	"foodworks.ml/m/internal/generated/graphql/model"
	"github.com/99designs/gqlgen/graphql"
)

func (r *carResolver) Users(ctx context.Context, obj *ent.Car) ([]*ent.User, error) {
	users, err := r.Client.User.
		Query().
		Where(user.HasCarsWith(car.IDEQ(obj.ID))).
		All(ctx)
	if err != nil {
		graphql.AddErrorf(ctx, "Error %d", err)
	}
	return users, err
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (int, error) {
	user, err := r.Client.User.
		Create().
		SetName(input.Name).
		SetAge(input.Age).
		Save(ctx)
	if err != nil {
		graphql.AddErrorf(ctx, "Error %d", err)
		return user.ID, err
	}
	err = r.Redis.Publish(ctx, "clients", user.ID).Err()
	return user.ID, err
}

func (r *queryResolver) Users(ctx context.Context, where *model.UserWhereInput) ([]*ent.User, error) {
	users, err := r.Client.User.
		Query().
		Where(user.NameEQ(where.Name)).
		All(ctx)
	if err != nil {
		graphql.AddErrorf(ctx, "Error %d", err)
	}
	return users, err
}

func (r *subscriptionResolver) OnUserCreated(ctx context.Context) (<-chan int, error) {
	channel := make(chan int, 1)
	go func() {
		sub := r.Redis.Subscribe(ctx, "clients")
		_, err := sub.Receive(ctx)
		if err != nil {
			return
		}
		ch := sub.Channel()
		for {
			select {
			case message := <-ch:
				id, err := strconv.Atoi(message.Payload)
				if err != nil {
					return
				}
				channel <- id
			// close when context done
			case <-ctx.Done():
				sub.Close()
				return
			}
		}
	}()
	return channel, nil
}

func (r *userResolver) Cars(ctx context.Context, obj *ent.User) ([]*ent.Car, error) {
	cars, err := r.Client.Car.
		Query().
		Where(car.HasOwnerWith(user.IDEQ(obj.ID))).
		All(ctx)
	if err != nil {
		graphql.AddErrorf(ctx, "Error %d", err)
	}
	return cars, err
}

// Car returns generated.CarResolver implementation.
func (r *Resolver) Car() generated.CarResolver { return &carResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type carResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
