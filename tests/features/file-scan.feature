Feature: testing virus scanning through rest API

    Scenario Outline: scan path for viruses
	Given I have a file path with contents <content> to scan with scanFile
	When I scan the file path for a virus
	Then I get a http status of <status> from scanFile

    Examples: virus_files
    | content                               | status |
    | "path=/clamav/tmp/ok/test.txt"        | "200"  |
    | "path=/clamav/tmp/virus/eicar.test"   | "406"  |
