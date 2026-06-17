# Usability

The syncer now has the feature that replying to private messages sends the reply to the sender on the other channel.

Also implement this for groups. When the robot forwards a group message to channel X, and when I reply to it, then my reply must go to the linked group on the other channel Y. 

If channel X is not linked, then the message must go to my private account on the other channel Y with a specific header stating that the group reply could not be completed.

Fix `README.md` accordingly if you make changes.
