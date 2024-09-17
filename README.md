![image](https://github.com/user-attachments/assets/a8d9e3a7-a8fe-4ee0-be4c-afc0c4913607)


### Project Structure

As the backend of the application is not large enough to break it down into smaller parts, it makes sense to implement all the functionality in a single file. This may seem unusual, but it really depends on the specific situation. Since the main focus of the backend part is on database connectivity and REST API controllers, it would be more convenient for developers to see the actual implementation in relation to how it is used. More specifically:

* `rest.go` - contains all the functionality
* `Dockerfile` & `docker-composel.yml` - it is the entrypoint of the infrastructure
* `integration_test.go` - test cases for the business logic
* `tests/...` - additional tests for the QA
* `.github/workflows/backend.yml` - CI for the GitHub Actions

### Abstract: Comprehensive Testing Suite for Robust Business Logic

In modern software development, ensuring the reliability and correctness of database operations is paramount. Our comprehensive testing suite is designed to rigorously validate core database functionalities, ensuring robustness, accuracy, and resilience. Leveraging `sqlmock` and `testify`, it simulates real-world scenarios, providing a controlled environment to test various outcomes without the need for a live database.

#### Key Features and Benefits:

1. **Full Coverage of Database Operations**:
   - Achieves _100%*_ coverage for business functionality, ensuring every line of code is tested and no unexpected failures occur. (*62.5% of overall code; 100% of business logic code)

2. **Simulation of Real-World Scenarios**:
   - Utilizes `sqlmock` to simulate real database interactions, including successful operations and various error conditions, allowing developers to anticipate and handle potential issues proactively.

3. **Efficient Error Detection and Handling**:
   - Tests for various error conditions, such as malformed queries and connection issues, ensuring the application can detect and handle errors efficiently, improving user experience and reducing downtime.

4. **Automated Testing and Continuous Integration**:
   - Every significant change that needs to be merged into the production environment will be automatically checked. The conscious integration process started instantly after the PR was opened to the main branch.
