document.addEventListener('DOMContentLoaded', () => {
    const searchBtn = document.getElementById('search-btn');
    const addressSearch = document.getElementById('address-search');
    const searchResults = document.getElementById('search-results');
    const generateIcsBtn = document.getElementById('generate-ics-btn');
    const uprnIcsInput = document.getElementById('uprn-ics-input');
    const icsLinkContainer = document.getElementById('ics-link-container');
    const icsActionsContainer = document.getElementById('ics-actions-container');

    const generateIcsLink = () => {
        const uprn = uprnIcsInput.value.trim();
        const sbcBaseUrl = 'https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days';

        // Clear previous content
        icsLinkContainer.innerHTML = '';
        icsActionsContainer.innerHTML = '';
        const upcomingCollectionsGrid = document.getElementById('upcoming-collections-grid');
        const upcomingCollectionsSection = document.getElementById('upcoming-collections');
        upcomingCollectionsGrid.innerHTML = '';
        upcomingCollectionsSection.style.display = 'none';

        if (uprn) {
            const icsLink = `${window.location.origin}/${uprn}/ics`;
            const encodedIcsLink = encodeURIComponent(icsLink);

            // 1. Display the main link
            const sbcCheckLink = `${sbcBaseUrl}?uprn=${uprn}&addressList=${uprn}&uprnSubmit=Yes`;
            icsLinkContainer.innerHTML = `<p><strong>Your calendar link is ready!</strong></p>
                                          <div class="link-wrapper">
                                            <a href="${icsLink}" target="_blank">${icsLink}</a>
                                          </div>
                                          <p>You can also <a href="${sbcCheckLink}" target="_blank">check your collection days on the SBC website</a> to verify the data.</p>
                                          <p>Now, copy the link or use the buttons below to add it to your calendar.</p>`;

            // 2. Create action buttons (Copy, Google, Outlook, Apple)
            const copyButton = document.createElement('button');
            copyButton.innerHTML = '📋'; // Clipboard emoji
            copyButton.title = 'Copy to Clipboard';
            copyButton.id = 'copy-ics-btn';
            copyButton.addEventListener('click', () => {
                navigator.clipboard.writeText(icsLink).then(() => {
                    copyButton.textContent = '✅';
                    setTimeout(() => { copyButton.innerHTML = '📋'; }, 2000);
                }, () => {
                    alert('Failed to copy');
                });
            });

            const googleCalendarLink = `https://calendar.google.com/calendar/render?cid=${encodedIcsLink}`;
            const outlookCalendarLink = `https://outlook.live.com/calendar/0/subscriptions/webcal/source?url=${encodedIcsLink}`;
            const appleCalendarLink = `webcal://${icsLink.replace(/^https?:\/\//, '')}`;

            const googleButton = `<a href="${googleCalendarLink}" target="_blank" class="calendar-btn google-btn">Add to Google Calendar</a>`;
            const outlookButton = `<a href="${outlookCalendarLink}" target="_blank" class="calendar-btn outlook-btn">Add to Outlook</a>`;
            const appleButton = `<a href="${appleCalendarLink}" class="calendar-btn apple-btn">Add to Apple Calendar</a>`;

            const buttonsWrapper = document.createElement('div');
            buttonsWrapper.className = 'buttons-wrapper';
            buttonsWrapper.innerHTML = `${googleButton}${outlookButton}${appleButton}`;

            icsActionsContainer.appendChild(copyButton);
            icsActionsContainer.appendChild(buttonsWrapper);

            // Fetch and display upcoming collections
            fetch(`${window.location.origin}/${uprn}/json`)
                .then(response => response.json())
                .then(data => {
                    if (data && data.collections && data.collections.length > 0) {
                        const flattenedCollections = [];
                        data.collections.forEach(collection => {
                            collection.CollectionDates.forEach(dateStr => {
                                const [year, month, day] = dateStr.split('-').map(Number);
                                flattenedCollections.push({
                                    date: new Date(year, month - 1, day),
                                    type: collection.type
                                });
                            });
                        });

                        // Sort by date ascending
                        flattenedCollections.sort((a, b) => a.date - b.date);

                        const table = document.createElement('table');
                        const thead = document.createElement('thead');
                        thead.innerHTML = '<tr><th>Date</th><th>Description</th></tr>';
                        table.appendChild(thead);

                        const tbody = document.createElement('tbody');
                        const options = { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' };
                        flattenedCollections.forEach(item => {
                            const row = document.createElement('tr');
                            const dateCell = document.createElement('td');
                            dateCell.textContent = item.date.toLocaleDateString(undefined, options);
                            const typeCell = document.createElement('td');
                            typeCell.textContent = item.type;
                            row.appendChild(dateCell);
                            row.appendChild(typeCell);
                            tbody.appendChild(row);
                        });
                        table.appendChild(tbody);

                        upcomingCollectionsGrid.appendChild(table);
                        upcomingCollectionsSection.style.display = 'block';
                    }
                })
                .catch(error => {
                    console.error('Error fetching upcoming collections:', error);
                });

        }
    };

    const performSearch = () => {
        const query = addressSearch.value.trim();
        if (query) {
            searchResults.textContent = 'Searching...';
            fetch(`search-address?q=${encodeURIComponent(query)}`)
                .then(response => response.json())
                .then(data => {
                    searchResults.innerHTML = '';
                    if (data && data.length > 0) {
                        const ul = document.createElement('ul');
                        data.forEach(item => {
                            const li = document.createElement('li');
                            li.innerHTML = `<strong>${item.address}</strong> (UPRN: ${item.uprn})`;
                            li.dataset.uprn = item.uprn;
                            li.addEventListener('click', () => {
                                uprnIcsInput.value = item.uprn;
                                generateIcsLink();
                                // Optional: scroll to the link container
                                document.getElementById('ics-generator').scrollIntoView({ behavior: 'smooth' });
                            });
                            ul.appendChild(li);
                        });
                        searchResults.appendChild(ul);
                    } else {
                        searchResults.textContent = 'No results found.';
                    }
                })
                .catch(error => {
                    console.error('Error fetching search results:', error);
                    searchResults.textContent = 'Failed to fetch results.';
                });
        }
    };

    searchBtn.addEventListener('click', performSearch);

    addressSearch.addEventListener('keypress', (event) => {
        if (event.key === 'Enter') {
            event.preventDefault();
            performSearch();
        }
    });

    generateIcsBtn.addEventListener('click', generateIcsLink);
});