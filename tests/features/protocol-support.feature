Feature: testing protocol support for the REST API

    Scenario Outline: scan files with different protocols
        Given I have a file with contents <content>
        When I scan the file using protocol <protocol>
        Then I get a http status of <status>
        And the protocol used is <protocol>

        Examples: clean_files
        | content       | protocol   | status |
        | "hello world" | "http1.1"  | "200"  |
        | "hello world" | "h2c"      | "200"  |
        | "hello world" | "https"    | "200"  |

        Examples: virus_files
        | content                                                                  | protocol   | status |
        | "X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*" | "http1.1"  | "406"  |
        | "X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*" | "h2c"      | "406"  |
        | "X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*" | "https"    | "406"  | 