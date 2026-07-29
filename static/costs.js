document.addEventListener('DOMContentLoaded', () => {
    const tree = document.getElementById('costs-tree');
    const totalCostEl = document.getElementById('total-cost');
    const currentYearEl = document.getElementById('current-year');
    const currentYear = new Date().getFullYear();
    currentYearEl.textContent = currentYear;

    const formatGBP = (amount) => `£${amount.toFixed(2)}`;

    const showMessage = (text, isError) => {
        tree.innerHTML = '';
        const p = document.createElement('p');
        if (isError) {
            p.className = 'error-text';
        }
        p.textContent = text;
        tree.appendChild(p);
    };

    fetch('/api/costs')
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            tree.innerHTML = '';

            if (!data || data.length === 0) {
                showMessage('No billing data available yet. Please check back later.', false);
                return;
            }

            // Group months by year, preserving the API's descending order.
            const years = [];
            const monthsByYear = new Map();
            data.forEach(item => {
                const year = String(item.year_month).slice(0, 4);
                if (!monthsByYear.has(year)) {
                    monthsByYear.set(year, []);
                    years.push(year);
                }
                monthsByYear.get(year).push(item);
            });
            years.sort((a, b) => b.localeCompare(a));

            let currentYearTotal = 0;

            years.forEach(year => {
                const months = monthsByYear.get(year);
                const yearTotal = months.reduce((sum, m) => sum + m.total_cost, 0);
                if (year === String(currentYear)) {
                    currentYearTotal = yearTotal;
                }

                const details = document.createElement('details');
                details.open = (year === String(currentYear));

                const summary = document.createElement('summary');
                const yearLabel = document.createElement('span');
                yearLabel.textContent = year;
                const yearTotalEl = document.createElement('span');
                yearTotalEl.className = 'year-total';
                yearTotalEl.textContent = formatGBP(yearTotal);
                summary.appendChild(yearLabel);
                summary.appendChild(yearTotalEl);
                details.appendChild(summary);

                const table = document.createElement('table');
                const thead = document.createElement('thead');
                thead.innerHTML = '<tr><th>Month</th><th>Total Cost</th><th>Note</th></tr>';
                table.appendChild(thead);

                const tbody = document.createElement('tbody');
                months.forEach(item => {
                    const row = document.createElement('tr');
                    const tdMonth = document.createElement('td');
                    tdMonth.textContent = item.year_month;
                    const tdCost = document.createElement('td');
                    tdCost.textContent = formatGBP(item.total_cost);
                    const tdNote = document.createElement('td');
                    tdNote.textContent = item.note || '';
                    row.appendChild(tdMonth);
                    row.appendChild(tdCost);
                    row.appendChild(tdNote);
                    tbody.appendChild(row);
                });
                table.appendChild(tbody);
                details.appendChild(table);

                tree.appendChild(details);
            });

            totalCostEl.textContent = formatGBP(currentYearTotal);
        })
        .catch(error => {
            console.error('Error fetching billing data:', error);
            showMessage('Failed to load billing data. Please try again later.', true);
        });
});
