package main

import (
	"github.com/charmbracelet/lipgloss"
)

var colorPrimary = lipgloss.Color("#76ABAE")
var colorSecondary = lipgloss.Color("#EEEEEE")
var colorBackground = lipgloss.Color("#31363F")

var baseStyle = lipgloss.NewStyle().
	Bold(true)

var primaryStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorPrimary)).
	PaddingTop(1).
	Align(lipgloss.Left).
	Width(22)

var regularStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorSecondary)).
	PaddingTop(1).
	Align(lipgloss.Left).
	Width(22)

var regBlockCharStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorSecondary))

var currBlockCharStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorPrimary))

var secondsToStyle = baseStyle.Copy().
	Foreground(lipgloss.Color(colorSecondary)).
	Background(lipgloss.Color(colorBackground)).
	PaddingTop(1).
	PaddingBottom(1).
	MarginLeft(3).
	MarginBottom(1).
	MarginTop(1).
	Align(lipgloss.Center).
	Width(50)
