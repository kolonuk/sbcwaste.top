This file contains instructions for AI agents working on this repository.

## Agent Instructions

*   **Goal:** Your primary goal is to assist users by completing coding tasks, such as solving bugs, implementing features, and writing tests.
*   **Tools:** You have access to a variety of tools to help you accomplish your goals. Use them wisely.
*   **Planning:** Always start by creating a solid plan. Explore the codebase, ask clarifying questions, and articulate your plan clearly.
*   **Verification:** Always verify your work. After every action that modifies the codebase, use a read-only tool to confirm that the action was successful.
*   **Testing:** Practice proactive testing. For any code change, attempt to find and run relevant tests. When practical, write a failing test first.
*   **Autonomy:** Strive to solve problems autonomously. However, don't hesitate to ask for help if you're stuck.
*   **Communication:** Keep the user informed of your progress. Use the `message_user` tool to provide updates.

## Adding a New Council

When adding a new council, please follow these steps:

1.  **Create a new scraper:** The scraper should be a new function in a new file in the `src/councils` directory.
2.  **Implement the `Council` interface:** The new scraper must implement the `Council` interface.
3.  **Add to the factory:** Add the new council to the `CouncilFactory` in `src/councils/factory.go`.
4.  **Add tests:** Add tests for the new scraper in a new test file in the `src/councils` directory.
5.  **Update documentation:** Update the `README.md` file to include the new council.
