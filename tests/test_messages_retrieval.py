import requests
import json

url = "http://localhost:8080/messages"

try:
    response = requests.get(f"{url}?OFFSET=12")
    response.raise_for_status()  # Check for HTTP errors
    messages = response.json()  # Parse the JSON response
    print("Messages:", messages)
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")
