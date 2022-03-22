**⚠️ Notice: This documentation is deprecated, please visit [docs.mantil.com](https://docs.mantil.com/concepts/cloud_development) to get the latest version!**

# Developing in The Cloud

In Mantil we chose to both develop and run production using cloud services. There is no copy of the cloud services for the local machine. Instead of trying to make copy of the cloud services for the local development we are making effort to get the feeling of the local development while using real services in the cloud. 

Under my experience having local development while using cloud services for production leads to supporting two environments. It is easy for trivial cases. But while the number of services or complexity of their interactions raise, as they always do in the real world, supporting two different environments becomes more and more painful. 

Developers like to have their own sandbox. With Mantil they have that private sandbox but instead of local it is using cloud resources. In this serverless world that means that development and production are essentially the same environment. It is not that one is using less capable servers, simplified network or something like that. All environments are the same!  
In Mantil we have concept of stage which is deployment of a project into cloud environment. By supporting infinite number of stages for each project development can be organized, besides private environment for each developer, into as many as needed integration, staging or show case stages. We make creating and deploying to the new stage a trivial step.

With Mantil you get all the benefits of the local development: isolated environment, instant feedback.   
While at the same time got other benefits:
 * no need to maintain two different environments
 * dev production and all the stages in between use exactly the same resources
 * Mantil handles everything no need to setup anything locally
 
But SAM, Serverless... have local development?  
Yes, but Mantil choose way to use cloud resources for both development and production. With the little change in mindset I believe that is long term right choice.

Questions I ask myself about team development environments:
 * how much time is needed to bootstrap new developer
   That developer can be a part time. To solve specific problem. Product manger who will edit texts. Designer who jump into project to make it usable. 
 * what is maintenance cost
   How much of his time developers spend building environment instead of business features.
 * how complex it is
   How many developers from the team are actually capable of extending development environment.
   What happens when initial developer leaves.
   

## How do I debug, can I set breakpoints in my function code?
No.  
When I started programming, many years ago, firing debugger and setting breakpoints was a way of life. That is especially convenient into Visual Studio or some other specialized IDE. The developers who are used to that kind of environments will feel unpleasant in any tool which doesn't support debug/breakpoint. I had the same feeling early in my career. After .Net I was developing in Ruby, Erlang, Go. In each of that environments I tried to setup some kind of breakpoint development style. That was short episodes and no one really useful. But I was not missing breakpoints. Breakpoint development is essentially a way to understand what is happening in the code. Once you have that mental model and can read code without need to fire debugger and go step by step you don't need it any more. Most of the experienced developers I know are asking for the breakpoint development style from the habit and from feeling insecure without it.  
Usually answer is that you don't need breakpoints just put some log lines to understand what is happening. My recommendation is to first build mental model about the code, then build test and to require that behavior from the code. Tests are repeatable and long lasting contract. Breakpoints are one time single mind explanation. 

Here is an ode to the debugging-less programming by the two legends in our filed. Quote is from "[The Best Programming Advice I Ever Got](http://www.informit.com/articles/article.aspx?p=1941206)" with Rob Pike:

> Ken taught me that thinking before debugging is extremely important. If you dive into the bug, you tend to fix the local issue in the code, but if you think about the bug first, how the bug came to be, you often find and correct a higher-level problem in the code that will improve the design and prevent further bugs.  
> I recognize this is largely a matter of style. Some people insist on line-by-line tool-driven debugging for everything. But I now believe that thinking—without looking at the code—is the best debugging tool of all, because it leads to better software.

<p align="right"> <a href="https://github.com/mantil-io/mantil/tree/master/docs#mantil-documentation">↵ Back to Documentation Home!</a></p>
