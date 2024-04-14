package main

import (
	"github.com/charmbracelet/lipgloss"
)

var colorPrimary = lipgloss.Color("#76ABAE")
var colorSecondary = lipgloss.Color("#EEEEEE")
var colorBackground = lipgloss.Color("#31363F")

var baseStyle = lipgloss.NewStyle().
	Bold(true)

var currStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorPrimary)).
	PaddingTop(1).
	Align(lipgloss.Left).
	Width(20)

var regularStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorSecondary)).
	PaddingTop(1).
	Align(lipgloss.Left).
	Width(22)

var timeToNextStyle = baseStyle.Copy().
	Foreground(lipgloss.Color("#607274")).
	MarginLeft(3).
	Align(lipgloss.Center)
