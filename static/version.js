document.addEventListener('DOMContentLoaded', () => {
    fetch('/api/version')
        .then((response) => response.json())
        .then((data) => {
            document.querySelectorAll('.app-version').forEach((el) => {
                el.textContent = `v${data.version}`;
            });
        })
        .catch((error) => {
            console.error('Error fetching version:', error);
        });
});
