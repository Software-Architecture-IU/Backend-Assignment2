import requests
import json
import time
from multiprocessing import Pool

BASE_URL = 'http://localhost:8080'
HEADERS = {'Content-Type': 'application/json'}

def post_message(text):
    payload = {"text": text}
    start_time = time.time()
    response = requests.post(f"{BASE_URL}/messages", headers=HEADERS, data=json.dumps(payload))
    end_time = time.time()
    return end_time - start_time, response

def get_messages(offset=0):
    start_time = time.time()
    response = requests.get(f"{BASE_URL}/messages?OFFSET={offset}", headers=HEADERS)
    end_time = time.time()
    return end_time - start_time, response

def get_message_count():
    start_time = time.time()
    response = requests.get(f"{BASE_URL}/messages/count", headers=HEADERS)
    end_time = time.time()
    return end_time - start_time, response

def test_single_message_one_user():
    times = []
    
    _, initial_count_response = get_message_count()
    initial_count = initial_count_response.json().get('int')
    
    post_time, _ = post_message("Test message 1")
    times.append(post_time)
    
    get_time, messages_response = get_messages()
    times.append(get_time)
    
    _, final_count_response = get_message_count()
    final_count = final_count_response.json().get('int')
    
    assert final_count == initial_count + 1, "Message count did not increase by 1"
    assert any(msg['text'] == "Test message 1" for msg in messages_response.json()), "Message not found in response"
    
    avg_time = sum(times) / len(times)
    print(f"Test #1 Average Roundtrip Time: {avg_time:.4f} seconds")
    return avg_time

def test_multiple_messages_one_user():
    times = []
    
    _, initial_count_response = get_message_count()
    initial_count = initial_count_response.json().get('int')
    
    for i in range(5):
        post_time, _ = post_message(f"Test message {i+1}")
        times.append(post_time)
    
    get_time, messages_response = get_messages()
    times.append(get_time)
    
    _, final_count_response = get_message_count()
    final_count = final_count_response.json().get('int')
    
    assert final_count == initial_count + 5, "Message count did not increase by 5"
    for i in range(5):
        assert any(msg['text'] == f"Test message {i+1}" for msg in messages_response.json()), f"Message {i+1} not found in response"
    
    avg_time = sum(times) / len(times)
    print(f"Test #2 Average Roundtrip Time: {avg_time:.4f} seconds")
    return avg_time

def post_message_multiprocessing(args):
    text, = args
    return post_message(text)

def test_single_message_multiple_users():
    times = []
    
    _, initial_count_response = get_message_count()
    initial_count = initial_count_response.json().get('int')
    
    with Pool(5) as p:
        results = p.map(post_message_multiprocessing, [("Test message",)] * 5)
    
    times.extend(rt for rt, _ in results)
    
    get_time, messages_response = get_messages()
    times.append(get_time)
    
    _, final_count_response = get_message_count()
    final_count = final_count_response.json().get('int')
    
    assert final_count == initial_count + 5, "Message count did not increase by 5"
    assert sum(msg['text'] == "Test message" for msg in messages_response.json()) == 5, "Not all messages found in response"
    
    avg_time = sum(times) / len(times)
    print(f"Test #3 Average Roundtrip Time: {avg_time:.4f} seconds")
    return avg_time

def test_multiple_messages_multiple_users():
    times = []
    
    _, initial_count_response = get_message_count()
    initial_count = initial_count_response.json().get('int')
    
    with Pool(5) as p:
        results = p.map(post_message_multiprocessing, [(f"Test message {i+1}",) for i in range(25)])
    
    times.extend(rt for rt, _ in results)
    
    get_time, messages_response = get_messages()
    times.append(get_time)
    
    _, final_count_response = get_message_count()
    final_count = final_count_response.json().get('int')
    
    assert final_count == initial_count + 25, "Message count did not increase by 25"
    for i in range(25):
        assert any(msg['text'] == f"Test message {i+1}" for msg in messages_response.json()), f"Message {i+1} not found in response"
    
    avg_time = sum(times) / len(times)
    print(f"Test #4 Average Roundtrip Time: {avg_time:.4f} seconds")
    return avg_time

if __name__ == "__main__":
    avg_times = []
    test_results = []

    avg_times.append(test_single_message_one_user())
    test_results.append({"test": "Test #1: Single Message One User", "avg_time": avg_times[-1]})

    avg_times.append(test_multiple_messages_one_user())
    test_results.append({"test": "Test #2: Multiple Messages One User", "avg_time": avg_times[-1]})

    avg_times.append(test_single_message_multiple_users())
    test_results.append({"test": "Test #3: Single Message Multiple Users", "avg_time": avg_times[-1]})

    avg_times.append(test_multiple_messages_multiple_users())
    test_results.append({"test": "Test #4: Multiple Messages Multiple Users", "avg_time": avg_times[-1]})

    overall_avg_time = sum(avg_times) / len(avg_times)
    print(f"Overall Average Roundtrip Time: {overall_avg_time:.4f} seconds")

    test_results.append({"test": "Overall", "avg_time": overall_avg_time})

    # Create loading_test_output.txt with Markdown formatted text
    with open("loading_test_output.txt", "w") as f:
        f.write("| Test Case | Average Roundtrip Time (seconds) |\n")
        f.write("|-----------|-----------------------------------|\n")
        for result in test_results:
            f.write(f"| {result['test']} | {result['avg_time']:.4f} ⚡|\n")
