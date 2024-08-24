# pce-label-migration
Copy ruleset and rules to migrate old labels to new labels. 
    
1. Export all rulesets and rules with workloader. Use option --no-href.      
2. Run the script "replace_ruleset" with -input <ruleset name>  
    This will generate a csv file to create the ruleset with new label both as the name and the ruleset scope. 
3. Import the ruleset with the workloader.  
4. Run the script "copy_rules". This script will generate a csv file to copy the following rules:  
    a. copy all rules in the old ruleset to the new ruleset 
    b. copy all rules in scopeless ruleset with specified application label and the label you want to migrate. New rules will have the replaced new label.  
    c. copy all extra-scope rules with specified application label and the label you want to migrate.   
4. Import the csv file with the workloader. 
