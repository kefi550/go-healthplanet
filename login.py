import os
from urllib.parse import urlencode, urlparse, parse_qs

import requests
from bs4 import BeautifulSoup


client_id = os.getenv("HEALTHPLANET_CLIENT_ID")
client_secret = os.getenv("HEALTHPLANET_CLIENT_SECRET")
login_id = os.getenv("HEALTHPLANET_LOGIN_ID")
login_password = os.getenv("HEALTHPLANET_LOGIN_PASSWORD")

redirect_url = "https://www.healthplanet.jp/success.html"
scope = "innerscan"
response_type = "code"
login_url = "https://www.healthplanet.jp/login_oauth.do"
auth_url = "https://www.healthplanet.jp/oauth/auth.do"
approval_url = "https://www.healthplanet.jp/oauth/approval.do"

session = requests.Session()

auth_query = urlencode(
    {
        "redirect_uri": redirect_url,
        "response_type": response_type,
        "client_id": client_id,
        "scope": scope,
    }
)

login_query = {
    "loginId": login_id,
    "passwd": login_password,
    "send": "1",
    "url": f"{auth_url}?{auth_query}",
}

response = session.post(login_url, data=login_query)
response_html = response.text

soup = BeautifulSoup(response_html, "html.parser")
oauth_token = soup.find("input", {"name": "oauth_token"}).get("value")

approval_query = {
    "approval": "true",
    "oauth_token": oauth_token,
}
response = session.post(approval_url, data=approval_query)
redirected_url = response.url
parsed_redirectd_url = urlparse(redirected_url)
query_string = parse_qs(parsed_redirectd_url.query)
code = query_string.get("code", [None])[0]

print(code)
