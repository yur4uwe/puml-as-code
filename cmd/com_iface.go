package main

type Command interface {
	Name() string
	Synopsis() string
	Usage() string
	Run(args []string) error
}
