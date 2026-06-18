# Setup

Linking groups is really painful. Revise the UI flow in the following manner:

1. Show a list of WhatsApp groups, prefixed by numbers.
2. Then ask which WhatsApp group to link. Accept a number matching the list, or `d` for done.
3. Then show a list of Signal groups, prefixed by numbers.
4. Then show the WhatsApp group that is to be linked, and ask for the number of the Signal group to link it to. Accept a number matching the list of Signal groups, or `n` for never mind.
5. Repeat the process until the user enters `d` in step 2.
