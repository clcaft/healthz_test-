package cmd

type ExampleCmdData struct {
	Id int64 `mapstructure:"iв" binding:"required"`
}
