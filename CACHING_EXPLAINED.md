# Caching in This Project

This document explains the caching implementation in this project and provides suggestions for potential improvements.

### How Caching Works

The application employs a dual-strategy caching system, designed to work efficiently in both local development and a deployed cloud environment.

1.  **Cache Abstraction**: The system is built around a `Cache` interface, which defines a standard set of operations (`Get`, `Set`, `Close`). This is a great practice as it decouples the application logic from the specific caching technology being used.

2.  **Two Implementations**:
    *   **SqliteCache**: For local development and testing (`APP_ENV=development` or `APP_ENV=test`), the application uses a local SQLite database file (`sbcwaste.db`). This is lightweight and avoids the need for developers to connect to cloud services for day-to-day work.
    *   **FirestoreCache**: In the production cloud environment, the application switches to Google Cloud Firestore. Firestore is a scalable, serverless NoSQL database that is well-suited for this type of caching, providing a robust solution for the deployed application.

3.  **How it's Used**: When the application needs to fetch waste collection data for a specific property (identified by a UPRN key):
    *   It first asks the cache for data associated with that key.
    *   If a valid, non-expired entry is found (a "cache hit"), the stored data is returned immediately, which is very fast.
    *   If the data is not in the cache or the cached entry has expired (a "cache miss"), the application scrapes the live data from the source website.
    *   This freshly scraped data is then stored in the cache with a set expiration time before being returned to the user. This ensures that the next request for the same property will be a cache hit.

### Suggestions for Improvement

The current implementation is solid, but we could make a few enhancements for better efficiency and reliability:

1.  **Use Firestore's Native TTL**: The `FirestoreCache` implementation currently checks for expiration within the application code. We could offload this work to Google Cloud by using **Firestore's native Time-to-Live (TTL) policies**. This would automatically delete expired cache documents, simplifying the code and potentially reducing the cost of read operations on expired data.
2.  **Graceful Cache Failure**: If the cache service (especially Firestore) were to become unavailable, the application should ideally continue to function by simply fetching fresh data from the source for every request. While this would be slower, it's better than failing completely. The error handling could be made more robust to ensure this graceful degradation.
3.  **Centralized Cache Management**: As has been implemented in the deployment workflow, there is now a process for managing the cache during deployments. Adding steps to the deployment workflow to report on and clear the cache is an excellent idea to ensure the application starts with a clean slate and avoids potential issues with outdated cache formats after an update.