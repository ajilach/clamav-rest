Feature: testing virus scanning through rest API

    Scenario Outline: scan files for viruses
	Given I have a file with contents <content> 
	When I scan the file for a virus
	Then I get a http status of <status>

        Examples: virus_files
        | content       | status |
        | "hello_world" | "200"   |
	| "X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*" | "406" |
