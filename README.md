# garden

Garden is a tool for orchestrating automated vulnerability scans for small
businesses.

## Disclaimer

Garden, and the accompanying research paper, were completed under a Research
Experiences for Undergraduates (REU) program, with funding from the National
Science Foundation (NSF).

## Usage

Garden comes with a man page to help explain some of the more complex aspects of
how it works. For a simple overview, please see the help menu.

```
Usage of garden:
  -category string
    	Categories to execute, separated by commas
  -context string
    	Docker context to use. Overrides -host
  -host string
    	Docker endpoint to use
  -ignore-hash
    	Skips checking module hashes
  -list
    	List loaded modules and their information
  -modargs string
    	Additional module-specific arguments
  -modules string
    	Directory to look for modules (default "./modules")
  -output string
    	Directory to output results (default "./reports")
  -single string
    	Individual modules to execute, separated by commas
  -target string
    	Host/IP to scan
  -v
    	Increase verbosity to info
  -version
    	Print the version and exit
  -vv
    	Increase verbosity to debug
```
