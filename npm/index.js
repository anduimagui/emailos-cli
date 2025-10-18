/**
 * mailos - EmailOS CLI
 * 
 * This package provides a Node.js wrapper for the mailos Go binary.
 * The actual functionality is implemented in Go and compiled to a native binary.
 * 
 * @see https://github.com/anduimagui/emailos
 */

module.exports = {
  name: 'mailos',
  version: require('./package.json').version,
  description: 'EmailOS - A standardized email client CLI'
};