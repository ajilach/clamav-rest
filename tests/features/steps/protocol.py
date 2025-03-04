from behave import when, then
import httpx

@when('I scan the file using protocol "{protocol}"')
def step_impl(context, protocol):
    files = {'file': context.file_contents}

    if protocol == "http1.1":
        # Use HTTPX with HTTP/1.1
        with httpx.Client(http1=True, http2=False) as client:
            r = client.post(f"{context.clamrest}/v2/scan", files=files)
    elif protocol == "h2c":
        # Use HTTPX with HTTP/2 prior knowledge mode
        with httpx.Client(http2=True, http1=False) as client:
            r = client.post(f"{context.clamrest}/v2/scan", files=files)
    elif protocol == "https":
        # Use HTTPX with HTTPS and HTTP/2
        https_url = context.clamrest.replace('http:', 'https:').replace(':9000', ':9443')
        with httpx.Client(verify=False, http2=True) as client:
            r = client.post(f"{https_url}/v2/scan", files=files)
    
    context.result = r

@then('the protocol used is "{protocol}"')
def step_impl(context, protocol):
    if protocol == "http1.1":
        assert context.result.http_version == "HTTP/1.1", f"Expected HTTP/1.1, got {context.result.http_version}"
    elif protocol in ["h2c", "https"]:
        assert context.result.http_version == "HTTP/2", f"Expected HTTP/2, got {context.result.http_version}"
    
    if protocol == "https":
        assert context.result.url.scheme == "https", "Expected HTTPS URL" 