package controllers

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewUserController),
	fx.Provide(NewCategoryController),
	fx.Provide(NewBrandController),
)
