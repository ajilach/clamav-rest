Feature: testing virus scanning through rest API

    Scenario Outline: scan path for viruses
	Given I have a path with contents <content> to scan with scanPath
	When I scan the path for a virus
	Then I get a http status of <status> from scanPath

    Examples: virus_files
    | content                            | status |
    | "path=/clamav/tmp/ok"              | "200"  |
    | "path=/clamav/tmp/virus"           | "406"  |
