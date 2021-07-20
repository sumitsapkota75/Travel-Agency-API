package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewUserRepository),
	fx.Provide(NewCategoryRepository),
	fx.Provide(NewBrandRepository),
	fx.Provide(NewProductRepository),
)
