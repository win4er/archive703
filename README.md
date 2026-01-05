# archive703
This is github repository for archive current knowledge and leaked test/exam papers and etc... With love from 703

Well, now I am trying to figure out what exactly structure of project I actually need

Let start from main functions:

1. Bot should store data
2. Bot should provide data in the simplest way
3. Users should been able to send request to store their data
4. Admin should been able to moderate all requests
5. Admin should been able to manage data in way he want
6. Data should have backup system or admin cant manage (delete) data

Additional functions:

7. Bot can have a GUI, idk, but i have already seen that WEB GUI in Telegram
8. Bot can have information about teachers, proffesors and etc
9. Bot can have a rating system that can be calculated based on how many information certain user has send and how others users liked/disliked it.
10. Bot can provide admins statistics and information about activity
11. Also bot can provide FAQ and can collect ideas from users for upgrading the project

Current thoughts about structure:

- ```admin.py``` is file for everything related to admin
- ```backend.py``` is file for main functions that common user has
- ```fronted.py``` is GUI file for bot

So, i think thats it...

