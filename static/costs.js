document.addEventListener('DOMContentLoaded', () => {
    const tbody = document.getElementById('costs-tbody');
    const totalCostEl = document.getElementById('total-cost');
    const currentYearEl = document.getElementById('current-year');
    const currentYear = new Date().getFullYear();
    currentYearEl.textContent = currentYear;

    fetch('/api/costs')
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            tbody.innerHTML = '';
            let totalCost = 0;

            if (!data || data.length === 0) {
                const row = document.createElement('tr');
                const cell = document.createElement('td');
                cell.colSpan = 2;
                cell.textContent = 'No billing data available yet. Please check back later.';
                row.appendChild(cell);
                tbody.appendChild(row);
                return;
            }

            data.forEach(item => {
                const row = document.createElement('tr');
                const tdMonth = document.createElement('td');
                tdMonth.textContent = item.year_month;
                const tdCost = document.createElement('td');
                tdCost.textContent = `£${item.total_cost.toFixed(2)}`;
                row.appendChild(tdMonth);
                row.appendChild(tdCost);
                tbody.appendChild(row);

                if (String(item.year_month).startsWith(currentYear)) {
                    totalCost += item.total_cost;
                }
            });

            totalCostEl.textContent = `£${totalCost.toFixed(2)}`;
        })
        .catch(error => {
            console.error('Error fetching billing data:', error);
            tbody.innerHTML = '';
            const row = document.createElement('tr');
            const cell = document.createElement('td');
            cell.colSpan = 2;
            cell.className = 'error-text';
            cell.textContent = 'Failed to load billing data. Please try again later.';
            row.appendChild(cell);
            tbody.appendChild(row);
        });
});
