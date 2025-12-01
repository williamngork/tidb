Currently, in info schema v2, we only load the special tables for  "admin repair table". 
So when CheckAndFetchRepairedTable() it cannot record and db info since the table list is empty.

Change the behaviour so that all tables are loaded for "admin repair table".
Ensures that tables in the repair table list are properly loaded in the info schema v2.
