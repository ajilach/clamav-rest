Feature: testing virus scanning through rest API

    Scenario Outline: scan body for viruses
	Given I have a body with contents <content>
	When I POST the content to /scanHandlerBody to scan the content for a virus
	Then I get a http status of <status> from /scanHandlerBody

    Examples: body_clean
    | content                                                                | status |
    | "hello_world"                                                          | "200"  |
	| "X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*" | "406"  |
