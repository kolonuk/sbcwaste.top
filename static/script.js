document.addEventListener('DOMContentLoaded', () => {
    const searchBtn = document.getElementById('search-btn');
    const addressSearch = document.getElementById('address-search');
    const searchResults = document.getElementById('search-results');
    const generateIcsBtn = document.getElementById('generate-ics-btn');
    const uprnIcsInput = document.getElementById('uprn-ics-input');
    const icsLinkContainer = document.getElementById('ics-link-container');
    const icsActionsContainer = document.getElementById('ics-actions-container');

    const uprnRegex = /^[0-9]{1,20}$/;

    const generateIcsLink = () => {
        const uprn = uprnIcsInput.value.trim();

        icsLinkContainer.innerHTML = '';
        icsActionsContainer.innerHTML = '';
        const upcomingCollectionsGrid = document.getElementById('upcoming-collections-grid');
        const upcomingCollectionsSection = document.getElementById('upcoming-collections');
        upcomingCollectionsGrid.innerHTML = '';
        upcomingCollectionsSection.classList.add('hidden');

        if (!uprn) {
            return;
        }

        if (!uprnRegex.test(uprn)) {
            const err = document.createElement('p');
            err.className = 'validation-error';
            err.textContent = 'Please enter a valid UPRN (digits only, up to 20 characters).';
            icsLinkContainer.appendChild(err);
            return;
        }

        const sbcBaseUrl = 'https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days';
        const icsLink = `${window.location.origin}/${uprn}/ics`;
        const encodedIcsLink = encodeURIComponent(icsLink);
        const sbcCheckLink = `${sbcBaseUrl}?uprn=${uprn}&addressList=${uprn}&uprnSubmit=Yes`;

        // Build ICS link display using DOM methods to avoid XSS
        const readyPara = document.createElement('p');
        const strong = document.createElement('strong');
        strong.textContent = 'Your calendar link is ready!';
        readyPara.appendChild(strong);
        icsLinkContainer.appendChild(readyPara);

        const linkWrapper = document.createElement('div');
        linkWrapper.className = 'link-wrapper';
        const linkEl = document.createElement('a');
        linkEl.href = icsLink;
        linkEl.target = '_blank';
        linkEl.rel = 'noopener noreferrer';
        linkEl.textContent = icsLink;
        linkWrapper.appendChild(linkEl);
        icsLinkContainer.appendChild(linkWrapper);

        const sbcPara = document.createElement('p');
        sbcPara.appendChild(document.createTextNode('You can also '));
        const sbcLink = document.createElement('a');
        sbcLink.href = sbcCheckLink;
        sbcLink.target = '_blank';
        sbcLink.rel = 'noopener noreferrer';
        sbcLink.textContent = 'check your collection days on the SBC website';
        sbcPara.appendChild(sbcLink);
        sbcPara.appendChild(document.createTextNode(' to verify the data.'));
        icsLinkContainer.appendChild(sbcPara);

        const helpPara = document.createElement('p');
        helpPara.textContent = 'Now, copy the link or use the buttons below to add it to your calendar.';
        icsLinkContainer.appendChild(helpPara);

        // Copy button
        const copyButton = document.createElement('button');
        copyButton.innerHTML = '&#x1F4CB;';
        copyButton.title = 'Copy to Clipboard';
        copyButton.id = 'copy-ics-btn';
        copyButton.addEventListener('click', () => {
            navigator.clipboard.writeText(icsLink).then(() => {
                copyButton.textContent = '✅';
                setTimeout(() => { copyButton.innerHTML = '&#x1F4CB;'; }, 2000);
            }, () => {
                alert('Failed to copy');
            });
        });

        // Calendar deep-link buttons
        const googleCalendarLink = `https://calendar.google.com/calendar/render?cid=${encodedIcsLink}`;
        const outlookCalendarLink = `https://outlook.live.com/calendar/0/subscriptions/webcal/source?url=${encodedIcsLink}`;
        const appleCalendarLink = `webcal://${icsLink.replace(/^https?:\/\//, '')}`;

        const googleBtn = document.createElement('a');
        googleBtn.href = googleCalendarLink;
        googleBtn.target = '_blank';
        googleBtn.rel = 'noopener noreferrer';
        googleBtn.className = 'calendar-btn google-btn';
        googleBtn.textContent = 'Add to Google Calendar';

        const outlookBtn = document.createElement('a');
        outlookBtn.href = outlookCalendarLink;
        outlookBtn.target = '_blank';
        outlookBtn.rel = 'noopener noreferrer';
        outlookBtn.className = 'calendar-btn outlook-btn';
        outlookBtn.textContent = 'Add to Outlook';

        const appleBtn = document.createElement('a');
        appleBtn.href = appleCalendarLink;
        appleBtn.className = 'calendar-btn apple-btn';
        appleBtn.textContent = 'Add to Apple Calendar';

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
                    upcomingCollectionsSection.classList.remove('hidden');
                }
            })
            .catch(error => {
                console.error('Error fetching upcoming collections:', error);
            });

        const buttonsWrapper = document.createElement('div');
        buttonsWrapper.className = 'buttons-wrapper';
        buttonsWrapper.appendChild(googleBtn);
        buttonsWrapper.appendChild(outlookBtn);
        buttonsWrapper.appendChild(appleBtn);

        icsActionsContainer.appendChild(copyButton);
        icsActionsContainer.appendChild(buttonsWrapper);
    };

    const performSearch = () => {
        const query = addressSearch.value.trim();
        if (!query) {
            return;
        }
        searchResults.textContent = 'Searching...';
        fetch(`search-address?q=${encodeURIComponent(query)}`)
            .then(response => response.json())
            .then(data => {
                searchResults.innerHTML = '';
                if (data && data.length > 0) {
                    const ul = document.createElement('ul');
                    data.forEach(item => {
                        const li = document.createElement('li');
                        const boldAddr = document.createElement('strong');
                        boldAddr.textContent = item.address;
                        li.appendChild(boldAddr);
                        li.appendChild(document.createTextNode(` (UPRN: ${item.uprn})`));
                        li.dataset.uprn = item.uprn;
                        li.addEventListener('click', () => {
                            uprnIcsInput.value = item.uprn;
                            generateIcsLink();
                            document.getElementById('upcoming-collections').scrollIntoView({ behavior: 'smooth' });
                        });
                        ul.appendChild(li);
                    });
                    searchResults.appendChild(ul);
                } else {
                    const msg = document.createElement('p');
                    msg.textContent = 'No results found. Try a full postcode (e.g. SN1 1AA) or just your street name.';
                    searchResults.appendChild(msg);
                }
            })
            .catch(error => {
                console.error('Error fetching search results:', error);
                searchResults.textContent = 'Failed to fetch results. Please try again.';
            });
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
