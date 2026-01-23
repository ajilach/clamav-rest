Feature: testing that a too large file gives 413

    Scenario Outline: scan larger file than allowed
	Given I have a too large file to scan with v2/scan 
	When I v2/scan a too large file
	Then I get a http status of <status> from v2/scan with a too large file

        Examples: virus_files
        | status |
        | "413"  |
