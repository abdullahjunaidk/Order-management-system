# Order Management System

## Project Overview 📝

This Order Management System is a microservices-based application ⚙️ designed to streamline e-commerce operations. It focuses on efficiently managing products 📦, inventory Stock📈, orders 🛍️, and payments 💳. The system is built with user authentication 🔑, supporting both vendor 🧑‍💼 and customer 🧑‍🧑‍🤝‍🧑 roles. It leverages asynchronous task processing ⏱️ for improved performance and scalability ☁️. Designed to be cloud-native 🌐, it utilizes Docker 🐳 for containerization and Consul 🧭 for service discovery.

## Services 🧩

The system is composed of distinct microservices, each handling a specific domain:

*   **Gateway Service 🚪**:
    *   Serves as the single entry point for all external API requests 🚦.
    *   Manages HTTP routing using Gin 🛤️.
    *   Enforces authentication and authorization via middleware 🛡️.
    *   Implements rate limiting using Redis to ensure API stability ⏳.
    *   Provides comprehensive API documentation with Swagger 📚.
    *   Communicates with backend services using gRPC 🗣️.

*   **Auth Service 🛡️**:
    *   Handles all aspects of user authentication and authorization 🔑.
    *   Manages user registration 📝, login ✅, logout 🚪, and secure password management (resetting and forgotten passwords 🔄).
    *   Generates and rigorously verifies JWT tokens for both access and refresh tokens 🗝️.
    *   Manages distinct user roles: customer and vendor 🧑‍💼🧑‍🧑‍🤝‍🧑.
    *   Persistently stores user data in MongoDB 💾.
    *   Publishes critical user-related events to RabbitMQ for asynchronous processing (e.g., user registration, account activation, password resets ✉️).

*   **Product Service 📦**:
    *   Manages the entire product catalog and detailed product information ℹ️.
    *   Empowers vendors to create ➕, update ✏️, delete 🗑️, and list products 📜.
    *   Stores all product-related information in MongoDB 💾.
    *   Dispatches product-related events via RabbitMQ (e.g., product creation, updates, deletions 📢).

*   **Inventory Service Stock📈**:
    *   Precisely manages product inventory levels 📊.
    *   Enables vendors to efficiently create ➕, update ✏️, delete 🗑️, and retrieve inventory details for their products 📦.
    *   Stores inventory data securely in MongoDB 💾.

*   **Order Service 🛍️**:
    *   Oversees the complete order lifecycle: creation ➕, retrieval 🔍, and cancellation ❌.
    *   Calculates order totals accurately and manages order statuses effectively 🔄.
    *   Stores comprehensive order information in MongoDB 💾.
    *   Broadcasts order-related events using RabbitMQ (e.g., order creation, cancellation, successful payments 📢).

*   **Payment Service 💳**:
    *   Facilitates secure payment processing using Stripe 💰.
    *   Generates unique payment links for each order 🔗.
    *   Processes real-time webhook events from Stripe for immediate payment confirmations and notifications of failures 📞.
    *   Interacts with Auth, Product, and Order services via high-performance gRPC 🗣️.
    *   Subscribes to relevant order and product events from RabbitMQ to orchestrate payment workflows and maintain data synchronization with Stripe 🔄.
    *   Sends timely email notifications for order confirmations and cancellations through the Mailer service ✉️.

*   **Mailer Service ✉️**:
    *   Dedicated to sending emails for crucial events: user registration, account activation, password resets, and order confirmations 📧.
    *   Utilizes Mailpit for efficient local email testing and SMTP for reliable real email delivery 🧪➡️📤.
    *   Subscribes to user-related events from RabbitMQ to trigger automated email sending processes 📬.

## Functionalities ✨

The Order Management System is packed with features, including:

*   **User Authentication and Authorization 🔑**:
    *   Streamlined user registration and secure account activation via email verification 📧✅.
    *   Secure user login and logout with JWT-based authentication 🔐🚪.
    *   Role-Based Access Control (RBAC) to manage permissions for customers and vendors 🧑‍💼🧑‍🧑‍🤝‍🧑.
    *   Password recovery and reset functionalities for account security 🔄.
    *   Refresh token mechanism for seamless access token renewal and persistent sessions 🔄.

*   **Product Management 📦**:
    *   Vendors can easily add new products with comprehensive details: name, description, category, and price ➕.
    *   Vendors have full control to update and remove their listed products ✏️🗑️.
    *   Vendors can efficiently list and manage their product offerings with pagination support 📜➡️.

*   **Inventory Management Stock📈**:
    *   Vendors can precisely control inventory levels, setting available and threshold quantities for each product 📊.
    *   Real-time inventory updates triggered by order placements and cancellations ensure accuracy 🔄.

*   **Order Management 🛍️**:
    *   Customers can create orders with multiple items, potentially from various vendors, in a single transaction 🛒.
    *   Orders are initiated in a "pending" status and transition to "paid" upon successful payment processing ⏳➡️✅.
    *   Customers can conveniently view their order details and cancel orders that are still pending 🔍❌.
    *   Unique payment links are dynamically generated for each order using Stripe for secure checkout 🔗💳.
    *   Real-time webhook handling for Stripe payment events ensures immediate order status updates and accurate records 📞🔄.

*   **Payment Processing 💳**:
    *   Seamless integration with Stripe for secure and reliable payment processing 🤝💰.
    *   Automated generation of Stripe Payment Links simplifies the checkout process for customers 🔗.
    *   Robust handling of Stripe webhook events keeps order statuses synchronized with payment events 📞🔄.
    *   Supports order cancellations and efficient payment refunds when necessary ❌↩️.

*   **Email Services ✉️**:
    *   Automated email notifications for user registration and account activation guide new users 📧✅.
    *   Password reset email functionality ensures users can recover access to their accounts securely 📧🔄.
    *   Order confirmation and cancellation emails keep customers informed about their transactions 📧🛍️❌.

*   **Monitoring and Tracing 🔭**:
    *   Distributed tracing with Jaeger provides deep insights into request flows across all services, aiding in monitoring and debugging 🔍.
    *   Centralized logging across all services using Logrus for efficient issue tracking and system analysis 🪵.
    *   Health check endpoints for every service enable proactive monitoring and ensure system health 🩺.

*   **Service Discovery and Configuration 🧭**:
    *   Consul facilitates dynamic service registration and discovery, enabling services to locate and communicate with each other automatically 📍🗣️.
    *   Environment variable-based configuration ensures flexibility and easy deployment across different environments ⚙️.

*   **API Security and Rate Limiting 🛡️⏳**:
    *   JWT-based authentication secures all API endpoints, protecting against unauthorized access 🔑.
    *   Rate limiting middleware in the Gateway service safeguards against abuse and ensures fair usage 🚦.

*   **API Documentation 📚**:
    *   Swagger documentation for the Gateway API provides interactive, up-to-date API exploration and integration resources for developers 📖.


## Docker Setup 🐳

To run the Order Management System using Docker Compose, ensure you have Docker and Docker Compose installed on your machine.

1.  **Clone the repository:**

    ```bash
    git clone <repository_url>
    cd <repository_directory>
    ```

2.  **Environment Variables:**

    *   Ensure you have configured all necessary environment variables. Example `.env` files are provided in each service directory (e.g., `.envs/auth/.env.example`).
    *   Copy `.env.example` files to `.env` and modify the values as needed for your setup.
    *   Pay special attention to database connection strings, RabbitMQ and Redis URIs, Stripe secrets, and JWT secrets.

3.  **Start the services with Docker Compose:**

    ```bash
    docker-compose up --build -d
    ```

    This command builds all the Docker images and starts all services in detached mode.

4.  **Access the applications:**

    *   **Gateway Service & Swagger UI**:  `http://localhost:8080/api/v1/swagger/index.html`
    *   **Mailpit UI**: `http://localhost:8025`
    *   **RabbitMQ Management UI**: `http://localhost:15672` (default credentials are `guest:guest`)
    *   **Jaeger UI**: `http://localhost:16686`
    *   **Consul UI**: `http://localhost:8500`
    *   **mongo-express UI**: `http://localhost:8081`

5.  **Stop the services:**

    ```bash
    docker-compose down
    ```

## Tech Stack 🛠️

*   **Backend Services**: Go 
*   **API Gateway Framework**: Gin 
*   **Inter-service Communication**: gRPC 
*   **Message Broker**: RabbitMQ 
*   **In-memory Data Store**: Redis 
*   **Database**: MongoDB 
*   **Payment Processing**: Stripe 
*   **Service Discovery**: Consul 
*   **Distributed Tracing**: Jaeger 
*   **Email Testing**: Mailpit 
*   **Containerization**: Docker

## License & Contributing 📜🧑‍🤝‍🧑

### License 📄

This project is licensed under the MIT License - see the [LICENSE](license) file for details.

### Contributing 🧑‍🤝‍🧑

Contributions are welcome! Please feel free to submit pull requests, report issues, or suggest new features to improve Gopher Social Backend.
