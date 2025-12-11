Feature: testing virus scanning with multiple files through rest API

    Scenario Outline: scan multiple files for viruses
	Given I have files with contents <content_a> and <content_b> to scan with v2 
	When I v2/scan the files for a virus
	Then I get a http status of <status> from v2/scan with multiple files

        Examples: virus_files
        | content_a     | content_b               | status |
        | "hello_world" | "hello_space"           | "200"  |
        | "hello_world" | "X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*" | "406" |
