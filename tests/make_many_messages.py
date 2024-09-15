import requests
import json

url = "http://localhost:8080/messages"

for i in range(1, 101):
    post_message = {
        "text": f"Hello, this is a message number {i}"
    }

    response = requests.post(url, json=post_message)
    print("Response Status:", response.status_code)
    print("Response Body:", response.text)

