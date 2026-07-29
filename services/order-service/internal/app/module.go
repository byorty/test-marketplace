package app

import "go.uber.org/fx"

var Module = fx.Options(

	ConfigModule,
	
	LoggerModule,

	DatabaseModule,

	AuthModule,

	RBACModule,

	RepositoryModule,
	
	ServiceModule,
	
	HandlerModule,

	ServerModule,
)